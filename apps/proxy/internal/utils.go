package internal

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isDirectIPAccess checks if the request is coming via direct IP access
func isDirectIPAccess(host string) bool {
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
