package gsh

import (
	"errors"
	"net/http"

	gocache "github.com/mrod502/go-cache"
)

var (
	ErrNotFound = errors.New("not found")
)

type Route struct {
	Ip    string
	Port  uint16
	Proto string
	cli   *http.Client
}

func (r *Route) Do(req *http.Request, replyto string) (*http.Response, error) {

	var req1 = new(http.Request)

	*req1 = *req

	req1.RemoteAddr = replyto

	return http.DefaultClient.Do(req1)
}

type RouteCache struct {
	v   *gocache.InterfaceCache
	cli *http.Client
}

func NewRouteCache(routes map[string]*Route) *RouteCache {

	return &RouteCache{}
}

func (r *RouteCache) Get(v string) (*Route, error) {
	val, err := r.v.Get(v)
	if err != nil {
		return nil, err
	}

	route, ok := val.(*Route)

	if !ok || route == nil {
		return nil, ErrNotFound
	}
	return route, nil
}

func (r *RouteCache) Set(k string, v *Route) {
	r.v.Set(k, v)
}

func (r *RouteCache) Do(req *http.Request) (*http.Response, error) {

	//v, err := r.Get(req.URL.EscapedPath())

	return nil, nil
}
