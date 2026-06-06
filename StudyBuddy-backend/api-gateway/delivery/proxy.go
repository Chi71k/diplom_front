package delivery

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	pkghttputil "studybuddy/backend/pkg/httputil"
)

// NewReverseProxy returns an http.Handler that forwards requests to baseURL with the
// original request path unchanged and the Host header rewritten to the upstream host.
func NewReverseProxy(baseURL string) (http.Handler, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, errors.New("empty upstream base URL")
	}
	target, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = proxyErrorHandler
	orig := proxy.Director
	proxy.Director = func(req *http.Request) {
		orig(req)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
	}
	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
	}
	return proxy, nil
}

func proxyErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		pkghttputil.Error(w, http.StatusGatewayTimeout, "gateway timeout")
		return
	}
	pkghttputil.Error(w, http.StatusBadGateway, "bad gateway")
}
