package gsh

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/logger"
)

type Router struct {
	routes     *gocache.StringCache // map[domain]localAddress
	cfg        *Config
	router     *mux.Router
	logger     *logger.Client
	logChan    chan []string
	routeCache *RouteCache
}

func NewRouter(cfg *Config, routes *gocache.StringCache) (r *Router, err error) {
	l, err := logger.NewClient(cfg.LoggerAddress, cfg.LogPrefix)
	if err != nil {
		return nil, err
	}
	r = &Router{
		routes:  routes,
		cfg:     cfg,
		router:  mux.NewRouter(),
		logger:  l,
		logChan: make(chan []string, 1024),
	}
	r.router.HandleFunc("/{path}", r.Forward)
	r.routeCache = NewRouteCache(1024, http.DefaultClient, routes)
	return r, nil
}

func (r *Router) Run() (err error) {
	go r.runLogger()
	go r.routeCache.Run()
	r.Info("START", fmt.Sprintf("Listening on port: %d", r.cfg.ServePort))
	if r.cfg.TLS {
		return http.ListenAndServeTLS(fmt.Sprintf(":%d", r.cfg.ServePort), r.cfg.CertFile, r.cfg.KeyFile, r.router)
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", r.cfg.ServePort), r.router)
}

func (r *Router) Info(inp ...string) {
	r.logChan <- inp
}

func (r *Router) Error(inp ...string) {
	r.logChan <- append([]string{"ERROR"}, inp...)

}
func (r *Router) runLogger() {
	for {
		l := <-r.logChan
		logger.Info(l...)
		r.logger.WriteLog(l...)

	}
}

func (r *Router) Forward(w http.ResponseWriter, req *http.Request) {
	host := r.routes.Get(req.URL.Hostname())
	r.Info("INCOMING", fmt.Sprintf("remote: %s | hostname: %s | to: %s | resource: %s", req.RemoteAddr, req.URL.Hostname(), host, req.URL.EscapedPath()))
	if host == "" {
		http.Error(w, "\n", http.StatusBadRequest)
		return
	}
	r.routeCache.Queue(NewHandleObject(req, w, host))

}
