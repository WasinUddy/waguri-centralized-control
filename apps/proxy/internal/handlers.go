package internal

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// serveMenu handles the menu page requests
func (s *Server) serveMenu(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Serving menu request",
		"method", r.Method,
		"host", r.Host,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.Header.Get("User-Agent"))

	// Handle API endpoint for services data
	if r.URL.Path == "/api/services" {
		s.serveServicesAPI(w, r)
		return
	}

	// Serve the HTML menu
	menuPath := filepath.Join(".", "menu.html")
	if _, err := os.Stat(menuPath); os.IsNotExist(err) {
		s.logger.Error("Menu file not found", "path", menuPath, "error", err)
		http.Error(w, "Menu file not found", http.StatusNotFound)
		return
	}

	// Read the HTML content
	tmplContent, err := os.ReadFile(menuPath)
	if err != nil {
		s.logger.Error("Error reading menu file", "path", menuPath, "error", err)
		http.Error(w, "Error reading menu file", http.StatusInternalServerError)
		return
	}

	// Serve the static HTML (services will be loaded via API)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write(tmplContent); err != nil {
		s.logger.Error("Error writing menu response:", err)
	}
}

// serveServicesAPI handles the API endpoint for services data
func (s *Server) serveServicesAPI(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Serving services API request",
		"method", r.Method,
		"host", r.Host,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr)

	services := s.generateServicesData()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(services); err != nil {
		s.logger.Error("Error encoding services response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleRequest is the main request handler that routes requests to appropriate handlers
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Log all incoming requests
	s.logger.Info("Incoming request",
		"method", r.Method,
		"host", r.Host,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.Header.Get("User-Agent"),
		"content_length", r.ContentLength)

	// Check if this is the menu host OR if accessing via IP (no Host header or IP format)
	if r.Host == s.cfg.Menu || isDirectIPAccess(r.Host) {
		s.logger.Info("Routing to menu handler",
			"host", r.Host,
			"is_menu_host", r.Host == s.cfg.Menu,
			"is_direct_ip", isDirectIPAccess(r.Host))
		s.serveMenu(w, r)
		duration := time.Since(startTime)
		s.logger.Info("Menu request completed",
			"duration_ms", duration.Milliseconds(),
			"host", r.Host,
			"path", r.URL.Path)
		return
	}

	// Check for proxy routes
	proxy, ok := s.proxyMap[r.Host]
	if !ok {
		s.logger.Info("No proxy route found, falling back to menu",
			"host", r.Host,
			"available_routes", len(s.proxyMap))
		s.serveMenu(w, r)
		duration := time.Since(startTime)
		s.logger.Info("Fallback to menu completed",
			"duration_ms", duration.Milliseconds(),
			"host", r.Host,
			"path", r.URL.Path)
		return
	}

	// Log proxy forward details
	var targetURL string
	for _, route := range s.cfg.Routes {
		if route.Host == r.Host {
			targetURL = route.Target
			break
		}
	}

	s.logger.Info("Proxying request to upstream",
		"method", r.Method,
		"source_host", r.Host,
		"target_url", targetURL,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
		"query", r.URL.RawQuery)

	// Create a custom response writer to capture status code
	wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	proxy.ServeHTTP(wrappedWriter, r)

	duration := time.Since(startTime)
	s.logger.Info("Proxy request completed",
		"method", r.Method,
		"source_host", r.Host,
		"target_url", targetURL,
		"path", r.URL.Path,
		"status_code", wrappedWriter.statusCode,
		"duration_ms", duration.Milliseconds(),
		"remote_addr", r.RemoteAddr)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
