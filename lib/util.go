package lib

import (
	"strings"
)

// singleJoiningSlach copied fro httputil package to support
// function newReverseProxy
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

// hostNameFromHostPort returns the host part from a string
// containing a "$host:$port" combination
// e.g.:
// 	host := hostNameFromHostPort("example.com:8080")
//	fmt.Println(host) // example.com
func hostNameFromHostPort(hostport string) string {
	parts := strings.Split(hostport, ":")
	return parts[0]
}
