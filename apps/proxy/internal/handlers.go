package internal

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

// serveMenu handles the menu page requests
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

// serveServicesAPI handles the API endpoint for services data
func (s *Server) serveServicesAPI(w http.ResponseWriter, _ *http.Request) {
	services := s.generateServicesData()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(services); err != nil {
		s.logger.Error("Error encoding services response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleRequest is the main request handler that routes requests to appropriate handlers
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Check if this is the menu host OR if accessing via IP (no Host header or IP format)
	if r.Host == s.cfg.Menu || isDirectIPAccess(r.Host) {
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
}
