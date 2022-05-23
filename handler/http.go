package handler

import (
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kernelschmelze/inbox/handler/limiter"
	"github.com/kernelschmelze/inbox/model"
	"github.com/kernelschmelze/inbox/store"

	log "github.com/kernelschmelze/pkg/logger"
)

type Handler struct {
	maxFileSize int64
	store       store.Store
	limiter     *limiter.IPRateLimiter
	plugins     []model.Plugin
	pGuard      sync.RWMutex
	process     chan *model.Inbox
	kill        chan struct{}
}

func New(maxFileSize int64, store store.Store) *Handler {

	handler := &Handler{
		maxFileSize: maxFileSize,
		store:       store,
		limiter:     limiter.NewIPRateLimiter(1, 5),
		process:     make(chan *model.Inbox, 100),
		kill:        make(chan struct{}),
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

		body, err := ioutil.ReadAll(file)
		if err != nil {
			log.Errorf("%T read data failed, err=%s", h, err)
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

		msg, err := model.NewInbox()
		if err != nil {
			log.Errorf("%T get new inbox failed, err=%s", h, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		filename := header.Filename
		if len(filename) < 2 {
			filename = ""
		}

		msg.From = from
		msg.Subject = subject
		msg.Filename = filename
		msg.Payload = body

		id := msg.ID

		if h.store != nil {

			id, err = h.store.Set(msg)

			if err != nil {
				log.Errorf("%T store data failed, err=%s", h, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

		}

		select {
		case h.process <- msg:
		case <-time.After(250 * time.Millisecond):
			log.Errorf("%T dispatch message failed, err=channel busy", h)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id))

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

			if msg == nil {
				continue
			}

			h.pGuard.RLock()

			for _, plugin := range h.plugins {
				plugin.Process(msg)
			}

			h.pGuard.RUnlock()

		}

	}

}
