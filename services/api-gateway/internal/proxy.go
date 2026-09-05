package internal

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"trading-platform/pkg/web"
)

func newProxy(rawURL string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	original := proxy.Director
	proxy.Director = func(r *http.Request) {
		incomingHost := r.Host
		original(r)
		r.Host = target.Host
		if incomingHost != "" {
			r.Header.Set("X-Forwarded-Host", incomingHost)
		}
		if r.TLS != nil {
			r.Header.Set("X-Forwarded-Proto", "https")
		} else if r.Header.Get("X-Forwarded-Proto") == "" {
			r.Header.Set("X-Forwarded-Proto", "http")
		}
	}
	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          32,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		web.WriteError(w, http.StatusBadGateway, web.CodeBadGateway, "Upstream unavailable")
	}
	proxy.FlushInterval = 100 * time.Millisecond
	return proxy, nil
}

func hopByHop(header http.Header) {
	for _, key := range []string{"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailers", "Transfer-Encoding", "Upgrade"} {
		header.Del(key)
	}
}
