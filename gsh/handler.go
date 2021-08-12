package gsh

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"go.uber.org/atomic"
)

func NewHandler(capacity uint32, client *http.Client, host string) (h *Handler) {
	var cli = new(http.Client)
	*cli = *client

	h = &Handler{
		queue:  make(chan *HandleObject, capacity),
		client: cli,
		host:   host,
		active: atomic.NewBool(true),
	}
	go h.processQueue()

	return
}

type Handler struct {
	queue  chan *HandleObject
	client *http.Client
	host   string
	active *atomic.Bool
}

func (h *Handler) processQueue() {
	for {
		h.handle(<-h.queue)
	}
}
func (h *Handler) handle(o *HandleObject) error {
	var req *http.Request = new(http.Request)
	*req = *o.R
	req.URL, _ = url.Parse(h.host + o.R.URL.EscapedPath())

	res, err := h.client.Do(req)
	if err != nil {
		http.Error(o.W, err.Error(), res.StatusCode)
		return err
	}
	b, _ := ioutil.ReadAll(res.Body)
	o.W.WriteHeader(res.StatusCode)
	_, err = o.W.Write(b)

	return err
}

func (h *Handler) Handle(o *HandleObject) {
	h.queue <- o
}
