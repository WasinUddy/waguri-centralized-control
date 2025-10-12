package internal

// ServiceInfo represents information about a service
type ServiceInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Icon        string `json:"icon"`
	Status      string `json:"status"`
	Category    string `json:"category"`
}

// generateServicesData creates a list of services from the proxy configuration
// Only includes services that have all required fields specified in the config
func (s *Server) generateServicesData() []ServiceInfo {
	var services []ServiceInfo

	for _, route := range s.cfg.Routes {
		// Skip routes that don't have all required fields
		if !s.isValidServiceConfig(route) {
			s.logger.Error("Skipping service route for", route.Host, "- missing required configuration fields")
			continue
		}

		service := ServiceInfo{
			Name:        route.Name,
			Description: route.Description,
			URL:         "http://" + route.Host,
			Icon:        route.Icon,
			Status:      "active",
			Category:    route.Category,
		}
		services = append(services, service)
	}

	return services
}

// isValidServiceConfig checks if a route has all required fields for service display
func (s *Server) isValidServiceConfig(route RoutesConfig) bool {
	if route.Name == "" {
		return false
	}
	if route.Description == "" {
		return false
	}
	if route.Icon == "" {
		return false
	}
	if route.Category == "" {
		return false
	}
	return true
}
