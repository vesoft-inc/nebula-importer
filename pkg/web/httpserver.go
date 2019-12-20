package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vesoft-inc/nebula-importer/pkg/cmd"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type WebServer struct {
	Port     int
	Callback string
	runner   *cmd.Runner
	server   *http.Server
}

func (w *WebServer) Start() {
	m := http.NewServeMux()

	m.HandleFunc("/submit", func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			w.submit(resp, req)
		}
	})

	m.HandleFunc("/stop", func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "PUT" {
			w.stop(resp, req)
		}
	})

	w.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", w.Port),
		Handler: m,
	}

	go w.listenAndServe()
}

func (w *WebServer) listenAndServe() {
	if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}
}

type body struct {
	FailedRows int64 `json:"failedRows"`
}

func (w *WebServer) callback(failedRows int64) {
	body := body{FailedRows: failedRows}
	b, err := json.Marshal(body)
	if err != nil {
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
		w.runner = &cmd.Runner{}
		go func() {
			defer req.Body.Close()
			var conf config.YAMLConfig
			if err := json.NewDecoder(req.Body).Decode(&conf); err != nil {
				http.Error(resp, err.Error(), http.StatusBadRequest)
			} else {
				w.runner.Run(conf)
				w.callback(w.runner.NumFailed)
				w.runner = nil
			}
		}()
		resp.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintln(resp, "OK"); err != nil {
			logger.Error(err)
		}
	}
}
