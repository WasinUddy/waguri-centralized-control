package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// serveMenu handles the menu page requests
func (s *Server) serveMenu(w http.ResponseWriter, r *http.Request) {
	s.logger.Info(fmt.Sprintf("Serving menu request - method=%s host=%s path=%s remote_addr=%s user_agent=%s",
		r.Method, r.Host, r.URL.Path, r.RemoteAddr, r.Header.Get("User-Agent")))

	// Handle API endpoint for services data
	if r.URL.Path == "/api/services" {
		s.serveServicesAPI(w, r)
		return
	}

	// Serve the HTML menu
	menuPath := filepath.Join(".", "menu.html")
	if _, err := os.Stat(menuPath); os.IsNotExist(err) {
		s.logger.Error(fmt.Sprintf("Menu file not found - path=%s error=%v", menuPath, err))
		http.Error(w, "Menu file not found", http.StatusNotFound)
		return
	}

	// Read the HTML content
	tmplContent, err := os.ReadFile(menuPath)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error reading menu file - path=%s error=%v", menuPath, err))
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
	s.logger.Info(fmt.Sprintf("Serving services API request - method=%s host=%s path=%s remote_addr=%s",
		r.Method, r.Host, r.URL.Path, r.RemoteAddr))

	services := s.generateServicesData()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(services); err != nil {
		s.logger.Error("Error encoding services response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// isWebSocketRequest checks if the request is a WebSocket upgrade request
func isWebSocketRequest(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("Upgrade")) == "websocket"
}

// handleWebSocketProxy handles WebSocket connection proxying
func (s *Server) handleWebSocketProxy(w http.ResponseWriter, r *http.Request, targetURL string) {
	// Parse target URL
	target, err := url.Parse(targetURL)
	if err != nil {
		s.logger.Error("Invalid target URL for WebSocket proxy:", err)
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		return
	}

	// Create WebSocket URL (convert http/https to ws/wss)
	wsScheme := "ws"
	if target.Scheme == "https" {
		wsScheme = "wss"
	}
	wsURL := fmt.Sprintf("%s://%s%s", wsScheme, target.Host, r.URL.Path)
	if r.URL.RawQuery != "" {
		wsURL += "?" + r.URL.RawQuery
	}

	// Upgrade the connection to WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for proxy
		},
	}

	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("Failed to upgrade client connection to WebSocket:", err)
		return
	}
	defer func() { _ = clientConn.Close() }()

	// Forward headers to the target
	headers := http.Header{}
	for key, values := range r.Header {
		// Skip connection-related headers that shouldn't be forwarded
		if strings.ToLower(key) == "connection" ||
			strings.ToLower(key) == "upgrade" ||
			strings.ToLower(key) == "sec-websocket-key" ||
			strings.ToLower(key) == "sec-websocket-version" ||
			strings.ToLower(key) == "sec-websocket-extensions" {
			continue
		}
		headers[key] = values
	}

	// Connect to the target WebSocket server
	targetConn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to connect to target WebSocket - url=%s error=%v", wsURL, err))
		_ = clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to connect to target"))
		return
	}
	defer func() { _ = targetConn.Close() }()

	// Start bidirectional message forwarding
	errChan := make(chan error, 2)

	// Forward messages from client to target
	go func() {
		for {
			messageType, message, err := clientConn.ReadMessage()
			if err != nil {
				errChan <- err
				return
			}

			err = targetConn.WriteMessage(messageType, message)
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Forward messages from target to client
	go func() {
		for {
			messageType, message, err := targetConn.ReadMessage()
			if err != nil {
				errChan <- err
				return
			}

			err = clientConn.WriteMessage(messageType, message)
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Wait for an error or connection close
	<-errChan
}

// handleRequest is the main request handler that routes requests to appropriate handlers
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Log all incoming requests
	s.logger.Info(fmt.Sprintf("Incoming request - method=%s host=%s path=%s remote_addr=%s user_agent=%s content_length=%d websocket=%t",
		r.Method, r.Host, r.URL.Path, r.RemoteAddr, r.Header.Get("User-Agent"), r.ContentLength, isWebSocketRequest(r)))

	// Check if this is the menu host OR if accessing via IP (no Host header or IP format)
	if r.Host == s.cfg.Menu || isDirectIPAccess(r.Host) {
		s.logger.Info(fmt.Sprintf("Routing to menu handler - host=%s is_menu_host=%t is_direct_ip=%t",
			r.Host, r.Host == s.cfg.Menu, isDirectIPAccess(r.Host)))
		s.serveMenu(w, r)
		duration := time.Since(startTime)
		s.logger.Info(fmt.Sprintf("Menu request completed - duration_ms=%d host=%s path=%s",
			duration.Milliseconds(), r.Host, r.URL.Path))
		return
	}

	// Check for redirect routes first
	if redirectURL, ok := s.redirectMap[r.Host]; ok {
		s.handleRedirect(w, r, redirectURL)
		duration := time.Since(startTime)
		s.logger.Info(fmt.Sprintf("Redirect completed - duration_ms=%d host=%s path=%s",
			duration.Milliseconds(), r.Host, r.URL.Path))
		return
	}

	// Check for proxy routes
	proxy, ok := s.proxyMap[r.Host]
	if !ok {
		s.logger.Info(fmt.Sprintf("No proxy route found, falling back to menu - host=%s available_routes=%d",
			r.Host, len(s.proxyMap)))
		s.serveMenu(w, r)
		duration := time.Since(startTime)
		s.logger.Info(fmt.Sprintf("Fallback to menu completed - duration_ms=%d host=%s path=%s",
			duration.Milliseconds(), r.Host, r.URL.Path))
		return
	}

	// Find the route configuration for this host
	var routeConfig *RoutesConfig
	for _, route := range s.cfg.Routes {
		if route.Host == r.Host {
			routeConfig = &route
			break
		}
	}

	// Safety check: if route is actually a redirect, handle it as redirect
	if routeConfig != nil && routeConfig.IsRedirect() {
		s.logger.Error(fmt.Sprintf("Route marked as proxy but is redirect - host=%s target=%s", r.Host, routeConfig.Target))
		s.handleRedirect(w, r, routeConfig.GetRedirectURL())
		duration := time.Since(startTime)
		s.logger.Info(fmt.Sprintf("Redirect completed (fallback) - duration_ms=%d host=%s path=%s",
			duration.Milliseconds(), r.Host, r.URL.Path))
		return
	}

	targetURL := ""
	if routeConfig != nil {
		targetURL = routeConfig.Target
	}

	// Check if this is a WebSocket request
	if isWebSocketRequest(r) {
		s.handleWebSocketProxy(w, r, targetURL)
		duration := time.Since(startTime)
		s.logger.Info(fmt.Sprintf("WebSocket proxy completed - duration_ms=%d host=%s path=%s",
			duration.Milliseconds(), r.Host, r.URL.Path))
		return
	}

	s.logger.Info(fmt.Sprintf("Proxying HTTP request to upstream - method=%s source_host=%s target_url=%s path=%s remote_addr=%s query=%s",
		r.Method, r.Host, targetURL, r.URL.Path, r.RemoteAddr, r.URL.RawQuery))

	// Create a custom response writer to capture status code
	wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	proxy.ServeHTTP(wrappedWriter, r)

	duration := time.Since(startTime)
	s.logger.Info(fmt.Sprintf("HTTP proxy request completed - method=%s source_host=%s target_url=%s path=%s status_code=%d duration_ms=%d remote_addr=%s",
		r.Method, r.Host, targetURL, r.URL.Path, wrappedWriter.statusCode, duration.Milliseconds(), r.RemoteAddr))
}

// handleRedirect handles HTTP redirects
func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request, redirectURL string) {
	// Construct the full redirect URL including path and query parameters
	fullRedirectURL := redirectURL + r.URL.Path
	if r.URL.RawQuery != "" {
		fullRedirectURL += "?" + r.URL.RawQuery
	}

	s.logger.Info(fmt.Sprintf("Redirecting request - source_host=%s source_path=%s redirect_url=%s remote_addr=%s method=%s",
		r.Host, r.URL.Path, fullRedirectURL, r.RemoteAddr, r.Method))

	// Send a 302 (Found) redirect
	http.Redirect(w, r, fullRedirectURL, http.StatusFound)
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
