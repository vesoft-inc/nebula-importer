package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/vesoft-inc/nebula-importer/pkg/cmd"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type WebServer struct {
	Port     int
	Callback string
	runner   *cmd.Runner
	server   *http.Server
	mux      sync.Mutex
}

func (w *WebServer) Start() {
	m := http.NewServeMux()

	m.HandleFunc("/submit", func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			w.mux.Lock()
			defer w.mux.Unlock()
			w.submit(resp, req)
		} else {
			http.Error(resp, "Invalid http method", http.StatusBadRequest)
		}
	})

	m.HandleFunc("/stop", func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "PUT" {
			w.mux.Lock()
			defer w.mux.Unlock()
			w.stop(resp, req)
		} else {
			http.Error(resp, "Invalid http method", http.StatusBadRequest)
		}
	})

	w.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", w.Port),
		Handler: m,
	}

	logger.Infof("Starting http server on %d", w.Port)
	w.listenAndServe()
}

func (w *WebServer) listenAndServe() {
	if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}
}

type respBody struct {
	ErrCode    int    `json:"errCode"`
	ErrMsg     string `json:"errMsg"`
	FailedRows int64  `json:"failedRows"`
}

func (w *WebServer) callback(body *respBody) {
	if b, err := json.Marshal(*body); err != nil {
		logger.Error(err)
	} else {
		_, err := http.Post(w.Callback, "application/json", bytes.NewBuffer(b))
		if err != nil {
			logger.Error(err)
		}
	}
}

func (w *WebServer) stop(resp http.ResponseWriter, req *http.Request) {
	if w.runner != nil {
		if w.runner.Readers == nil {
			http.Error(resp, "Retry stop again", http.StatusBadRequest)
			return
		}
		for _, r := range w.runner.Readers {
			r.Stop()
		}
	}
	resp.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintln(resp, "OK"); err != nil {
		logger.Error(err)
	}
}

func (w *WebServer) submit(resp http.ResponseWriter, req *http.Request) {
	if w.runner != nil {
		msg := "There some running tasks, please wait a minute"
		logger.Error(msg)
		http.Error(resp, msg, http.StatusTooManyRequests)
	} else {
		if req.Body == nil {
			http.Error(resp, "nil request body", http.StatusBadRequest)
			return
		}
		defer req.Body.Close()
		var conf config.YAMLConfig
		if err := json.NewDecoder(req.Body).Decode(&conf); err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}
		w.runner = &cmd.Runner{}
		go func() {
			w.runner.Run(&conf)
			if w.runner.Error() != nil {
				logger.Error(w.runner.Error())
				w.callback(&respBody{
					ErrCode: 1,
					ErrMsg:  w.runner.Error().Error(),
				})
			} else {
				w.callback(&respBody{
					ErrCode:    0,
					FailedRows: w.runner.NumFailed,
				})
			}
			w.runner = nil
		}()
		resp.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintln(resp, "OK"); err != nil {
			logger.Error(err)
		}
	}
}
