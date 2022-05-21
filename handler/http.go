package handler

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/kernelschmelze/inbox/handler/limiter"
	"github.com/kernelschmelze/inbox/model"

	log "github.com/kernelschmelze/pkg/logger"

	"github.com/google/uuid"
)

type Handler struct {
	documentRoot string
	maxFileSize  int64
	limiter      *limiter.IPRateLimiter
	plugins      []model.Plugin
	pGuard       sync.RWMutex
	process      chan model.Inbox
	kill         chan struct{}
}

func New(documentRoot string, maxFileSize int64) *Handler {

	handler := &Handler{
		documentRoot: documentRoot,
		maxFileSize:  maxFileSize,
		limiter:      limiter.NewIPRateLimiter(1, 5),
		process:      make(chan model.Inbox, 100),
		kill:         make(chan struct{}),
	}
	go handler.run()

	return handler

}

func (h *Handler) Close() {
	close(h.kill)
	h.pGuard.RLock()
	for _, plugin := range h.plugins {
		plugin.Close()
	}
	h.pGuard.RUnlock()
}

func (h *Handler) AddPlugin(p model.Plugin) {
	h.pGuard.Lock()
	h.plugins = append(h.plugins, p)
	h.pGuard.Unlock()
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if h.limiter != nil {
		limiter := h.limiter.GetLimiter(r.RemoteAddr)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
	}

	switch r.Method {

	case http.MethodPost:

		if path := strings.Trim(r.URL.Path, "/"); path != "inbox" {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		defer file.Close()

		if h.maxFileSize > 0 && header.Size > h.maxFileSize {
			http.Error(w, "file size exceeds the limit", http.StatusNotAcceptable)
			return
		}

		id, err := uuid.NewRandom()
		if err != nil {
			log.Errorf("%T get uuid failed, err=%s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		body, err := ioutil.ReadAll(file)
		if err != nil {
			log.Errorf("%T read data failed, err=%s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		from := r.FormValue("from")
		subject := r.FormValue("subject")

		if len(from) > 50 {
			from = from[:50]
		}

		if len(subject) > 80 {
			subject = subject[:80]
		}

		if len(body) == 0 && len(subject) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		path := path.Join(h.documentRoot, id.String())
		if err = mkdir(path); err != nil {
			log.Errorf("%T path failed, err=%s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Errorf("%T open file failed, err=%s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		filename := header.Filename
		if len(filename) < 2 {
			filename = ""
		}

		defer output.Close()

		var b64 bool

		if b64 = !isPrintable(body); b64 == true {
			body = []byte(base64.StdEncoding.EncodeToString(body))
		}

		msg := model.Inbox{
			Time:     time.Now(),
			ID:       id.String(),
			From:     from,
			Subject:  subject,
			Filename: filename,
			Base64:   b64,
			Payload:  body,
		}

		encoder := json.NewEncoder(output)
		if err = encoder.Encode(msg); err != nil {
			log.Errorf("%T encode failed, err=%s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		select {
		case h.process <- msg:
		case <-time.After(1 * time.Second):
			log.Errorf("%T dispatch message failed, err=channel busy", h)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id.String()))

	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}

}

func (h *Handler) run() {

	for {

		select {

		case <-h.kill:
			return

		case msg := <-h.process:

			h.pGuard.RLock()

			for _, plugin := range h.plugins {
				plugin.Process(msg)
			}

			h.pGuard.RUnlock()

		}

	}

}

func mkdir(file string) error {
	path := file
	index := strings.LastIndex(file, "/")
	if index > 0 {
		path = file[0:index]
	}
	err := os.MkdirAll(path, 0700)
	return err
}

func isPrintable(bytes []byte) bool {
	for i := range bytes {
		c := bytes[i]
		if (c < 32 || c > 126) && (c != 9 && c != 10 && c != 13) {
			return false
		}
	}
	return true
}
