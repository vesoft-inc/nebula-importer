package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type WebServer struct {
	stopCh       chan bool
	shutdownCh   chan bool
	HttpSettings *config.HttpSettings
	server       *http.Server
}

func (w *WebServer) Start() {
	w.stopCh = make(chan bool)
	w.shutdownCh = make(chan bool)
	m := http.NewServeMux()
	m.HandleFunc("/stop", func(resp http.ResponseWriter, r *http.Request) {
		if _, err := resp.Write([]byte("OK")); err != nil {
			logger.Error(err)
		}
		w.stopCh <- true
	})

	port := *w.HttpSettings.Port

	w.server = &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: m}

	go w.listenAndServe()
	go w.waitForShutdown()
}

func (w *WebServer) listenAndServe() {
	if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}
}

func (w *WebServer) waitForShutdown() {
	<-w.shutdownCh
	close(w.stopCh)
	if err := w.server.Shutdown(context.Background()); err != nil {
		logger.Error(err)
	}
	logger.Infof("Shutdown http server listened on %d", *w.HttpSettings.Port)
}

type body struct {
	FailedRows int64 `json:"failedRows"`
}

func (w *WebServer) Shutdown(failedRows int64) {
	if w.HttpSettings.Callback != nil {
		body := body{failedRows: failedRows}
		b, err := json.Marshal(body)
		if err != nil {
			logger.Error(err)
		} else {
			_, err := http.Post(*w.HttpSettings.Callback, "application/json", bytes.NewBuffer(b))
			if err != nil {
				logger.Error(err)
			}
		}
	}
	w.shutdownCh <- true
}

func (w *WebServer) Stop(a []interface{}) {
	if s, ok := <-w.stopCh; ok && s {
		for _, s := range a {
			if ss, ok := s.(base.Stoppable); !ok {
				logger.Error("Error type cast to stoppable interface")
			} else {
				ss.Stop()
			}
		}
	}
}
