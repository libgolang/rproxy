package lib

import (
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/libgolang/log/v2"
)

type backend struct {
	last    int
	proxies []*httputil.ReverseProxy
}

func newBackend(cfg *Config, urls []string) *backend {
	proxies := make([]*httputil.ReverseProxy, 0, 0)
	for _, u := range urls {
		remote, err := url.Parse(u)
		if err != nil {
			log.Fatal("%s", err)
		}
		proxies = append(proxies, newReverseProxy(cfg, remote))
	}

	b := &backend{
		last:    len(proxies),
		proxies: proxies,
	}
	return b
}

func (o *backend) Balance(r *http.Request) *httputil.ReverseProxy {
	if o.last > 0 {
		i := rand.Intn(o.last)
		log.Info("rand: %d", i)
		return o.proxies[i]
	} else {
		return o.proxies[0]
	}
}

// newReverseProxy copied from httputil package to customize
// creation of *httputil.ReverseProxy
func newReverseProxy(cfg *Config, target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", target.Host)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
	}

	var transport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          *cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   *cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
	}
}
