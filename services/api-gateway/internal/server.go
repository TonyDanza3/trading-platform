package internal

import (
	"io"
	"net/http"
	"os"
	"time"

	"trading-platform/pkg/httpx"
	"trading-platform/pkg/web"
)

type Server struct {
	cfg     Config
	routes  []Route
	proxies map[string]http.Handler
	ready   *http.Client
}

func New(cfg Config) (*Server, error) {
	routes := cfg.Routes()
	proxies := make(map[string]http.Handler, len(routes))
	for _, route := range routes {
		if _, ok := proxies[route.Target]; ok {
			continue
		}
		proxy, err := newProxy(route.Target)
		if err != nil {
			return nil, err
		}
		proxies[route.Target] = proxy
	}
	return &Server{
		cfg:     cfg,
		routes:  routes,
		proxies: proxies,
		ready:   &http.Client{Timeout: 2 * time.Second},
	}, nil
}

func (s *Server) Handler() http.Handler {
	return httpx.Wrap(http.HandlerFunc(s.serve))
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/health":
		web.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	case "/ready":
		s.readyCheck(w, r)
		return
	case "/openapi.yaml":
		s.serveOpenAPI(w)
		return
	case "/swagger", "/swagger/":
		serveSwagger(w)
		return
	}

	route, ok := matchRoute(r.URL.Path, s.routes)
	if !ok {
		web.WriteError(w, http.StatusNotFound, web.CodeNotFound, "No route for this path")
		return
	}
	proxy := s.proxies[route.Target]
	hopByHop(r.Header)
	proxy.ServeHTTP(w, r)
}

func (s *Server) readyCheck(w http.ResponseWriter, r *http.Request) {
	upstreams := []string{s.cfg.AuthURL, s.cfg.UserURL, s.cfg.MarketURL}
	for _, base := range upstreams {
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, base+"/health", nil)
		if err != nil {
			web.WriteError(w, http.StatusServiceUnavailable, web.CodeUnavailable, "Upstream unavailable")
			return
		}
		if id := r.Header.Get("X-Request-Id"); id != "" {
			req.Header.Set("X-Request-Id", id)
		}
		resp, err := s.ready.Do(req)
		if err != nil {
			web.WriteError(w, http.StatusServiceUnavailable, web.CodeUnavailable, "Upstream unavailable")
			return
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			web.WriteError(w, http.StatusServiceUnavailable, web.CodeUnavailable, "Upstream unavailable")
			return
		}
	}
	web.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) serveOpenAPI(w http.ResponseWriter) {
	data, err := os.ReadFile(s.cfg.OpenAPIPath)
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "OpenAPI spec not found")
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func serveSwagger(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8"/>
  <title>Trading Platform API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({ url: "/openapi.yaml", dom_id: "#swagger-ui" });
  </script>
</body>
</html>`))
}
