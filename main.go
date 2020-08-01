package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/libgolang/log/v2"
	"github.com/libgolang/props"
)

func main() {
	if props.IsSet("d") {
		log.SetLevel(log.LevelTrace)
	}
	configFile := props.GetProp("config")

	cfg, err := readConfig(configFile)
	if err != nil {
		log.Fatal("%s", err)
	}

	http.DefaultTransport.(*http.Transport).MaxIdleConns = *cfg.MaxIdleConns
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = *cfg.MaxIdleConnsPerHost

	http.Handle("/", newHandler(cfg))
	listenTo := fmt.Sprintf("%s:%d", *cfg.Ip, *cfg.Port)
	log.Infof("Starting server %s", listenTo)
	err = http.ListenAndServe(listenTo, nil)
	if err != nil {
		log.Fatal("%s", err)
	}
}

type handler struct {
	m map[string]*backend
}

func newHandler(cfg *config) *handler {
	m := make(map[string]*backend)
	for _, p := range cfg.Proxies {
		log.Info("Loading %s -> %s Proxy", p.DomainName, p.Remote)
		m[p.DomainName] = newBackend(p.Remote)
	}
	return &handler{m}
}

func (o *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hostName := hostNameFromHostPort(r.Host)
	if proxy, ok := o.m[hostName]; ok {
		backend := proxy.Balance()
		backend.ServeHTTP(w, r)
	} else {
		log.Warnf("Hostname %s not found", hostName)
		for k, v := range o.m {
			log.Info("%s -> %v", k, v)
		}

		w.WriteHeader(404)
		w.Write([]byte("Error 404 (Not Found)"))
	}
}

func readConfig(file string) (*config, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file %s: %s", file, err)
	}
	cfg := &config{}
	if err := toml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("Error reading config file %s: %s", file, err)
	}

	if cfg.Ip == nil {
		*cfg.Ip = "0.0.0.0"
	}
	if cfg.MaxIdleConns == nil {
		*cfg.MaxIdleConns = 10000
	}
	if cfg.MaxIdleConnsPerHost == nil {
		*cfg.MaxIdleConnsPerHost = 10000
	}
	if cfg.Port == nil {
		*cfg.Port = 8080
	}
	return cfg, nil
}

type backend struct {
	last    int
	proxies []*httputil.ReverseProxy
}

func newBackend(urls []string) *backend {
	proxies := make([]*httputil.ReverseProxy, 0, 0)
	for _, u := range urls {
		remote, err := url.Parse(u)
		if err != nil {
			log.Fatal("%s", err)
		}
		proxies = append(proxies, newReverseProxy(remote))
	}

	b := &backend{
		last:    len(proxies),
		proxies: proxies,
	}
	return b
}

func (o *backend) Balance() *httputil.ReverseProxy {
	if o.last > 0 {
		i := rand.Intn(o.last)
		log.Info("rand: %d", i)
		return o.proxies[i]
	} else {
		return o.proxies[0]
	}
}

func newReverseProxy(target *url.URL) *httputil.ReverseProxy {
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
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func hostNameFromHostPort(hostport string) string {
	parts := strings.Split(hostport, ":")
	return parts[0]
}

type config struct {
	Ip                  *string       `toml:"ip"`
	Port                *int          `toml:"port"`
	MaxIdleConns        *int          `toml:"maxIdleConns"`        //10000
	MaxIdleConnsPerHost *int          `toml:"maxIdleConnsPerHost"` // 10000
	Proxies             []configProxy `toml:"proxy"`
}

type configProxy struct {
	DomainName string   `toml:"domainName"` // example.com
	Remote     []string `toml:"remote"`     // example.com
}
