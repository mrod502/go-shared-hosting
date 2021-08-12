package gsh

import (
	"errors"
	"net/http"
	"sync"

	gocache "github.com/mrod502/go-cache"
)

var (
	ErrKey  = errors.New("key not found")
	ErrFull = errors.New("route cache is full")
)

type HandleObject struct {
	R    *http.Request
	W    http.ResponseWriter
	Host string
}

func NewHandleObject(r *http.Request, w http.ResponseWriter, host string) *HandleObject {
	return &HandleObject{R: r, W: w, Host: host}
}

type RouteCache struct {
	m    *sync.RWMutex
	v    map[string]*Handler
	reqs chan *HandleObject
}

//Set a value
func (r *RouteCache) Set(k string, v *Handler) {
	r.m.Lock()
	defer r.m.Unlock()
	r.v[k] = v

}

//Set a value
func (r *RouteCache) Get(k string) *Handler {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.v[k]

}
func (r *RouteCache) do(h *HandleObject) {
	r.m.RLock()
	defer r.m.RUnlock()
	r.v[h.Host].Handle(h)
}

func (r *RouteCache) GetKeys() (out []string) {
	r.m.RLock()
	defer r.m.RUnlock()
	out = make([]string, len(r.v))
	var i int
	for k := range r.v {
		out[i] = k
	}
	return out
}

func NewRouteCache(capacity uint32, client *http.Client, routes *gocache.StringCache) (r *RouteCache) {
	r = &RouteCache{m: new(sync.RWMutex), v: make(map[string]*Handler), reqs: make(chan *HandleObject, capacity)}
	for _, v := range routes.GetKeys() {
		r.Set(v, NewHandler(capacity, client, routes.Get(v)))
	}
	return r
}
func (r *RouteCache) Run() {
	for {
		r.do(<-r.reqs)
	}
}

func (r *RouteCache) Queue(o *HandleObject) error {
	if cap(r.reqs) == len(r.reqs) {
		return ErrFull
	}
	r.reqs <- o

	return nil
}
