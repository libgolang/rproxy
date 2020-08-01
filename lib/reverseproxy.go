package lib

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/libgolang/log"
)

type ReverseProxyBackend interface {
	Balance(r *http.Request) *httputil.ReverseProxy
}

type ReverseProxyBackendMatcher interface {
	MatchBackend(r *http.Request) ReverseProxyBackend
}

type ReverseProxy struct {
	Server         *http.Server
	mux            *http.ServeMux
	handler        *reverseProxyHandler
	BackendMatcher ReverseProxyBackendMatcher
}

type reverseProxyHandler struct {
	backendMatcher ReverseProxyBackendMatcher
}

type defaultBackendMatcher struct {
	m map[string]*backend
}

func NewReverseProxy(cfg *Config) *ReverseProxy {
	listenTo := fmt.Sprintf("%s:%d", *cfg.Ip, *cfg.Port)
	log.Infof("Starting server %s", listenTo)

	// reverse proxy handler
	m := make(map[string]*backend)
	for _, p := range cfg.Proxies {
		log.Infof("Loading %s -> %s Proxy", p.DomainName, p.Remote)
		m[p.DomainName] = newBackend(cfg, p.Remote)
	}

	// mux
	handler := &reverseProxyHandler{}
	mux := http.NewServeMux()
	mux.Handle("/", handler)

	// server
	server := &http.Server{}
	server.Addr = listenTo
	server.Handler = mux
	return &ReverseProxy{
		Server:         server,
		mux:            mux,
		handler:        handler,
		BackendMatcher: &defaultBackendMatcher{m: m},
	}
}

func (o *ReverseProxy) ListenAndServe() error {
	o.handler.backendMatcher = o.BackendMatcher
	return o.Server.ListenAndServe()
}

func (o *reverseProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := o.backendMatcher.MatchBackend(r)
	if backend != nil {
		proxy := backend.Balance(r)
		proxy.ServeHTTP(w, r)
	} else {
		// 404 not found
		hostName := hostNameFromHostPort(r.Host)
		log.Errorf("No backend found for %s : %s", hostName, r.URL)
		w.WriteHeader(404)
		w.Write([]byte("Error 404 (Not Found)"))
	}
}

func (o *defaultBackendMatcher) MatchBackend(r *http.Request) ReverseProxyBackend {
	hostName := hostNameFromHostPort(r.Host)
	if backend, ok := o.m[hostName]; ok {
		return backend
	}
	return nil
}
