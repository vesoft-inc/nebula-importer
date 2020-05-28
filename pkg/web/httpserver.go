package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/vesoft-inc/nebula-importer/pkg/cmd"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type WebServer struct {
	Port     int
	Callback string
	server   *http.Server
	taskMgr  *taskMgr
	mux      sync.Mutex
}

var taskId uint64 = 0

func (w *WebServer) newTaskId() string {
	w.mux.Lock()
	defer w.mux.Unlock()
	tid := taskId
	taskId++
	return fmt.Sprintf("%d", tid)
}

func (w *WebServer) Start() {
	m := http.NewServeMux()
	w.taskMgr = newTaskMgr()

	m.HandleFunc("/submit", func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			w.submit(resp, req)
		} else {
			w.badRequest(resp, "HTTP method must be POST")
		}
	})

	m.HandleFunc("/stop", func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "PUT" {
			w.stop(resp, req)
		} else {
			w.badRequest(resp, "HTTP method must be PUT")
		}
	})

	m.HandleFunc("/tasks", func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			keys := w.taskMgr.keys()
			var tasks struct {
				Tasks []string `json:"tasks"`
			}
			tasks.Tasks = keys
			if b, err := json.Marshal(tasks); err != nil {
				w.badRequest(resp, err.Error())
			} else {
				resp.WriteHeader(http.StatusOK)
				if _, err = resp.Write(b); err != nil {
					logger.Error(err)
				}
			}
		} else {
			w.badRequest(resp, "HTTP method must be GET")
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

type errResult struct {
	ErrCode int    `json:"errCode"`
	ErrMsg  string `json:"errMsg"`
}

type task struct {
	errResult
	TaskId string `json:"taskId"`
}

type respBody struct {
	task
	FailedRows int64 `json:"failedRows"`
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

func (w *WebServer) stopRunner(taskId string) {
	runner := w.taskMgr.get(taskId)
	if runner == nil {
		return
	}

	for _, r := range runner.Readers {
		r.Stop()
	}

	logger.Infof("Task %s stopped.", taskId)
}

func (w *WebServer) stop(resp http.ResponseWriter, req *http.Request) {
	if req.Body == nil {
		w.badRequest(resp, "nil request body")
		return
	}
	defer req.Body.Close()

	var task task
	if err := json.NewDecoder(req.Body).Decode(&task); err != nil {
		w.badRequest(resp, err.Error())
		return
	}

	if strings.ToLower(task.TaskId) == "all" {
		for _, k := range w.taskMgr.keys() {
			w.stopRunner(k)
		}
	} else {
		w.stopRunner(task.TaskId)
	}

	resp.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintln(resp, "OK"); err != nil {
		logger.Error(err)
	}
}

func (w *WebServer) badRequest(resp http.ResponseWriter, msg string) {
	resp.WriteHeader(http.StatusOK)
	t := errResult{
		ErrCode: 1,
		ErrMsg:  msg,
	}

	if b, err := json.Marshal(t); err != nil {
		logger.Error(err)
	} else {
		resp.WriteHeader(http.StatusOK)
		if _, err = resp.Write(b); err != nil {
			logger.Error(err)
		}
	}
}

func (w *WebServer) submit(resp http.ResponseWriter, req *http.Request) {
	if req.Body == nil {
		w.badRequest(resp, "nil request body")
		return
	}
	defer req.Body.Close()

	var conf config.YAMLConfig
	if err := json.NewDecoder(req.Body).Decode(&conf); err != nil {
		w.badRequest(resp, err.Error())
		return
	}

	if err := conf.ValidateAndReset(""); err != nil {
		w.badRequest(resp, err.Error())
		return
	}

	runner := &cmd.Runner{}
	tid := w.newTaskId()
	w.taskMgr.put(tid, runner)
	t := task{
		errResult: errResult{ErrCode: 0},
		TaskId:    tid,
	}

	go func(tid string) {
		runner.Run(&conf)
		var body respBody
		rerr := runner.Error()
		if rerr != nil {
			err, _ := rerr.(errors.ImporterError)
			logger.Error(err)
			body = respBody{
				task: task{
					errResult: errResult{
						ErrCode: err.ErrCode,
						ErrMsg:  err.ErrMsg.Error(),
					},
					TaskId: tid,
				},
			}
		} else {
			body = respBody{
				task:       t,
				FailedRows: runner.NumFailed,
			}
		}
		w.callback(&body)
		w.taskMgr.del(tid)
	}(tid)

	if b, err := json.Marshal(t); err != nil {
		w.badRequest(resp, err.Error())
	} else {
		resp.WriteHeader(http.StatusOK)
		if _, err := resp.Write(b); err != nil {
			logger.Error(err)
		}
	}
}
