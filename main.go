package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/libgolang/log"
	"github.com/libgolang/props"
)

func main() {
	if props.IsSet("d") {
		log.SetLevel(log.LevelTrace)
	}
	configFile := props.GetProp("config")

	cfg, err := readConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handler(cfg))
	listenTo := fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port)
	log.Infof("Starting server %s", listenTo)
	err = http.ListenAndServe(listenTo, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(cfg *config) func(http.ResponseWriter, *http.Request) {
	m := make(map[string]*httputil.ReverseProxy)
	for _, p := range cfg.Proxies {
		log.Info("Loading %s -> %s Proxy", p.DomainName, p.Remote)
		remote, err := url.Parse(p.Remote)
		if err != nil {
			log.Fatal(err)
		}
		m[p.DomainName] = newReverseProxy(remote)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if proxy, ok := m[r.Host]; ok {
			log.Info(r.URL)
			proxy.ServeHTTP(w, r)
		} else {
			log.Warnf("Hostname %s not found", r.Host)
			w.WriteHeader(404)
			w.Write([]byte("Error 404 (Not Found)"))
		}
	}
}

func readConfig(file string) (*config, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file %s: %s", file, err)
	}
	cfg := &config{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("Unable to parse JSON for file %s: %s", file, err)
	}
	return cfg, nil
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

type config struct {
	Ip      string        `json:"ip"`
	Port    int           `json:"port"`
	Proxies []configProxy `json:"proxies"`
}

type configProxy struct {
	DomainName string `json:"domainName"` // example.com
	Remote     string `json:"remote"`     // example.com
}
