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

	// Handler for all requests
	mux.HandleFunc("/", s.handleRequest)

	server := &http.Server{
		Addr:    s.cfg.Listen,
		Handler: mux,
	}

	s.logger.Info("Proxy server starting on", s.cfg.Listen)
	s.logger.Info("Menu available at:", "http://"+s.cfg.Menu)
	s.logger.Info("Menu also accessible via direct IP access")
	return server.ListenAndServe()
}
