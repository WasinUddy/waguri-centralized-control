package internal

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"waguri-centralized-control/packages/go-utils/telemetry"
)

type Server struct {
	cfg      *ProxyConfig
	logger   *telemetry.Logger
	proxyMap map[string]*httputil.ReverseProxy
}

func NewServer(cfg *ProxyConfig, logger *telemetry.Logger) *Server {
	s := &Server{
		cfg:      cfg,
		logger:   logger,
		proxyMap: make(map[string]*httputil.ReverseProxy),
	}

	for _, route := range cfg.Routes {
		targetURL, err := url.Parse(route.Target)
		if err != nil {
			logger.Error("Skipping invalid target URL for host", route.Host, ":", err)
			continue
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		s.proxyMap[route.Host] = proxy
		logger.Info("Registered route:", route.Host, "->", targetURL.String())
	}

	return s
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Single handler for all requests
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy, ok := s.proxyMap[r.Host]
		if !ok {
			http.Error(w, "Unknown host: "+r.Host, http.StatusNotFound)
			return
		}
		s.logger.Info("Proxying", r.Method, "request for", r.Host, "to upstream")
		proxy.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:    s.cfg.Listen,
		Handler: mux,
	}

	s.logger.Info("Proxy server starting on", s.cfg.Listen)
	return server.ListenAndServe()
}
