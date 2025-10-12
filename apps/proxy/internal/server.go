package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"waguri-centralized-control/packages/go-utils/telemetry"
)

type Server struct {
	cfg      *ProxyConfig
	logger   *telemetry.Logger
	proxyMap map[string]*httputil.ReverseProxy
}

type ServiceInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Icon        string `json:"icon"`
	Status      string `json:"status"`
	Category    string `json:"category"`
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

func (s *Server) serveMenu(w http.ResponseWriter, r *http.Request) {
	// Handle API endpoint for services data
	if r.URL.Path == "/api/services" {
		s.serveServicesAPI(w, r)
		return
	}

	// Serve the HTML menu
	menuPath := filepath.Join(".", "menu.html")
	if _, err := os.Stat(menuPath); os.IsNotExist(err) {
		http.Error(w, "Menu file not found", http.StatusNotFound)
		return
	}

	// Read the HTML content
	tmplContent, err := os.ReadFile(menuPath)
	if err != nil {
		http.Error(w, "Error reading menu file", http.StatusInternalServerError)
		return
	}

	// Serve the static HTML (services will be loaded via API)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write(tmplContent); err != nil {
		s.logger.Error("Error writing menu response:", err)
	}
}

func (s *Server) serveServicesAPI(w http.ResponseWriter, _ *http.Request) {
	services := s.generateServicesData()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(services); err != nil {
		s.logger.Error("Error encoding services response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (s *Server) generateServicesData() []ServiceInfo {
	var services []ServiceInfo

	for _, route := range s.cfg.Routes {
		// Use configured values or generate fallbacks
		name := route.Name
		if name == "" {
			name = s.generateServiceName(route.Host)
		}

		description := route.Description
		if description == "" {
			description = s.generateServiceDescription(route.Host, route.Target)
		}

		icon := route.Icon
		if icon == "" {
			icon = "monitor" // default Lucide icon
		}

		category := route.Category
		if category == "" {
			category = "Services"
		}

		service := ServiceInfo{
			Name:        name,
			Description: description,
			URL:         "http://" + route.Host,
			Icon:        icon,
			Status:      "active",
			Category:    category,
		}
		services = append(services, service)
	}

	return services
}

func (s *Server) generateServiceName(host string) string {
	// Convert hostname to friendly name
	nameMap := map[string]string{
		"api.waguri.san": "API Gateway",
		"app1.nas.happy": "Application 1",
		"app2.nas.happy": "Application 2",
	}

	if name, exists := nameMap[host]; exists {
		return name
	}

	// Generate name from hostname
	return "Service: " + host
}

func (s *Server) generateServiceDescription(host, target string) string {
	// Generate description based on host and target
	descMap := map[string]string{
		"api.waguri.san": "Main API endpoint for all backend services",
		"app1.nas.happy": "Primary application service running on NAS",
		"app2.nas.happy": "Secondary application service",
	}

	if desc, exists := descMap[host]; exists {
		return desc
	}

	return "Service running on " + target
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Handler for all requests
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is the menu host OR if accessing via IP (no Host header or IP format)
		if r.Host == s.cfg.Menu || s.isDirectIPAccess(r.Host) {
			s.serveMenu(w, r)
			return
		}

		// Check for proxy routes
		proxy, ok := s.proxyMap[r.Host]
		if !ok {
			// If no proxy route found and not menu host, serve menu as fallback
			s.serveMenu(w, r)
			return
		}
		s.logger.Info("Proxying", r.Method, "request for", r.Host, "to upstream", "requested from", r.RemoteAddr)
		proxy.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:    s.cfg.Listen,
		Handler: mux,
	}

	s.logger.Info("Proxy server starting on", s.cfg.Listen)
	s.logger.Info("Menu available at:", "http://"+s.cfg.Menu)
	s.logger.Info("Menu also accessible via direct IP access")
	return server.ListenAndServe()
}

// isDirectIPAccess checks if the request is coming via direct IP access
func (s *Server) isDirectIPAccess(host string) bool {
	// If no host header or empty host
	if host == "" {
		return true
	}

	// Check if host is an IP address (contains only digits, dots, and colons for IPv6)
	// Simple check for IP-like patterns
	if isIPAddress(host) {
		return true
	}

	// Check if host includes port with IP (e.g., "192.168.1.1:80")
	if colonIndex := len(host) - 1; colonIndex > 0 {
		for i := colonIndex; i >= 0; i-- {
			if host[i] == ':' {
				potentialIP := host[:i]
				if isIPAddress(potentialIP) {
					return true
				}
				break
			}
		}
	}

	return false
}

// isIPAddress checks if a string looks like an IP address
func isIPAddress(s string) bool {
	if s == "" {
		return false
	}

	// Simple IPv4 check - contains only digits and dots
	digitCount := 0
	dotCount := 0

	for _, char := range s {
		if char >= '0' && char <= '9' {
			digitCount++
		} else if char == '.' {
			dotCount++
		} else {
			// For IPv6 or other formats, we'll also accept colons
			if char != ':' {
				return false
			}
		}
	}

	// Basic IPv4 format check (should have 3 dots and some digits)
	return dotCount == 3 && digitCount > 0
}
