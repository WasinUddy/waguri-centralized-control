package internal

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"waguri-centralized-control/packages/go-utils/telemetry"
)

type Server struct {
	cfg         *ProxyConfig
	logger      *telemetry.Logger
	proxyMap    map[string]*httputil.ReverseProxy
	redirectMap map[string]string
}

func NewServer(cfg *ProxyConfig, logger *telemetry.Logger) *Server {
	s := &Server{
		cfg:         cfg,
		logger:      logger,
		proxyMap:    make(map[string]*httputil.ReverseProxy),
		redirectMap: make(map[string]string),
	}

	for _, route := range cfg.Routes {
		if route.IsRedirect() {
			// Handle redirect routes
			redirectURL := route.GetRedirectURL()
			s.redirectMap[route.Host] = redirectURL
			logger.Info("Registered redirect route:", route.Host, "->", redirectURL)
		} else {
			// Handle proxy routes
			targetURL, err := url.Parse(route.Target)
			if err != nil {
				logger.Error("Skipping invalid target URL for host", route.Host, ":", err)
				continue
			}
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			s.proxyMap[route.Host] = proxy
			logger.Info("Registered proxy route:", route.Host, "->", targetURL.String())
		}
	}

	return s
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Serve static assets from ./static at the /static/ path
	staticDir := "./static"
	fileServer := http.FileServer(http.Dir(staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/waguri.ico")
	})

	// Handler for all other requests
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
