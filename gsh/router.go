package gsh

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/go-shared-hosting/obj"
	"github.com/mrod502/logger"
)

type Router struct {
	routes     map[string]*Route
	cfg        *Config
	router     *mux.Router
	logger     logger.Client
	routeCache *gocache.InterfaceCache
	cli        *obj.ClientCache
	replyTo    string
}

func NewRouter(cfg *Config, routes map[string]*Route) (r *Router, err error) {
	l, err := logger.NewClient(cfg.ClientConfig)
	if err != nil {
		return nil, err
	}
	r = &Router{
		routes: routes,
		cfg:    cfg,
		router: mux.NewRouter(),
		logger: l,
		cli:    obj.NewClientCache(1<<13, time.Minute*3),
	}

	for path, route := range r.routes {
		r.router.HandleFunc(path, r.genHandleFunc(route))
	}
	r.routeCache = gocache.NewInterfaceCache()
	return r, nil
}

func (r *Router) Run() (err error) {
	err = r.logger.Connect()
	if err != nil {
		return err
	}

	r.Info("START", fmt.Sprintf("Listening on port: %d", r.cfg.ServePort))
	if r.cfg.TLS {
		return http.ListenAndServeTLS(fmt.Sprintf(":%d", r.cfg.ServePort), r.cfg.CertFile, r.cfg.KeyFile, r.router)
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", r.cfg.ServePort), r.router)
}

func (r *Router) Info(inp ...string) {
	r.logger.Write(append([]string{"INF"}, inp...)...)
}

func (r *Router) Error(inp ...string) {
	r.logger.Write(append([]string{"ERR"}, inp...)...)

}

func (r *Router) genHandleFunc(route *Route) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {
		authorized, err := r.cli.Authorize(req)
		if !authorized {
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}
			http.Error(w, errMsg, http.StatusTooManyRequests)
			r.Error("fwd", req.RemoteAddr, err.Error())
			return
		}
		res, err := route.Do(req, r.replyTo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			r.Error("from:", req.RemoteAddr, "to:", fmt.Sprintf("%s:%d", route.Ip, route.Port), req.URL.EscapedPath())
			return
		}
		b, _ := io.ReadAll(res.Body)

		for k := range res.Header {
			w.Header().Set(k, res.Header.Get(k))
		}
		_, err = w.Write(b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			r.Error("from:", req.RemoteAddr, "to:", fmt.Sprintf("%s:%d", route.Ip, route.Port), req.URL.EscapedPath())
			return
		}
	}
}
