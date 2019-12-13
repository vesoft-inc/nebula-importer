package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

var StopCh chan bool
var ShutdownCh chan bool

func Start(port int) {
	StopCh = make(chan bool)
	ShutdownCh = make(chan bool)
	m := http.NewServeMux()
	m.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("OK")); err != nil {
			logger.Error(err)
		}
		StopCh <- true
	})

	s := http.Server{Addr: fmt.Sprintf(":%d", port), Handler: m}

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()

	go func() {
		<-ShutdownCh
		close(StopCh)
		s.Shutdown(context.Background())
		logger.Infof("Shutdown http server listened on %d", port)
	}()
}
