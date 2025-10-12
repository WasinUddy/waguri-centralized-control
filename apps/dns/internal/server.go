package internal

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"waguri-centralized-control/packages/go-utils/config"
	"waguri-centralized-control/packages/go-utils/telemetry"

	dnslib "github.com/miekg/dns"
)

type Server struct {
	cfg       *config.Config
	logger    *telemetry.Logger
	dnsServer *dnslib.Server
	// Compiled regex patterns for wildcard domains
	wildcardPatterns map[*regexp.Regexp]string
}

func NewServer(cfg *config.Config, logger *telemetry.Logger) *Server {
	server := &Server{
		cfg:              cfg,
		logger:           logger,
		wildcardPatterns: make(map[*regexp.Regexp]string),
	}

	// Compile wildcard patterns
	server.compileWildcardPatterns()

	return server
}

// compileWildcardPatterns converts wildcard domain patterns to regex
func (s *Server) compileWildcardPatterns() {
	for domain, ip := range s.cfg.Domains {
		if strings.Contains(domain, "*") {
			// Convert wildcard pattern to regex
			// *.waguri.san becomes ^[^.]+\.waguri\.san$
			regexPattern := strings.ReplaceAll(domain, ".", "\\.")
			regexPattern = strings.ReplaceAll(regexPattern, "*", "[^.]+")
			regexPattern = "^" + regexPattern + "$"

			if compiled, err := regexp.Compile(regexPattern); err == nil {
				s.wildcardPatterns[compiled] = ip
				s.logger.Info("Compiled wildcard pattern:", domain, "->", regexPattern, "resolves to", ip)
			} else {
				s.logger.Error("Failed to compile wildcard pattern for", domain, err)
			}
		}
	}
}

// findDomainMatch checks for exact match first, then wildcard patterns
func (s *Server) findDomainMatch(domain string) (string, bool) {
	// First try exact match
	if ip, ok := s.cfg.Domains[domain]; ok {
		return ip, true
	}

	// Then try wildcard patterns
	for pattern, ip := range s.wildcardPatterns {
		if pattern.MatchString(domain) {
			return ip, true
		}
	}

	return "", false
}

func (s *Server) handleDNS(w dnslib.ResponseWriter, r *dnslib.Msg) {
	m := new(dnslib.Msg)
	m.SetReply(r)
	m.Authoritative = true

	// Log the incoming query details
	clientAddr := w.RemoteAddr()
	s.logger.Info("Received DNS query from", clientAddr, "- ID:", r.Id, "Questions:", len(r.Question))

	for _, q := range r.Question {
		s.logger.Info("Query:", q.Name, "Type:", dnslib.TypeToString[q.Qtype], "Class:", dnslib.ClassToString[q.Qclass])

		name := strings.ToLower(strings.TrimSuffix(q.Name, "."))

		// Lookup using both exact and wildcard matching
		ip, ok := s.findDomainMatch(name)
		if ok {
			rr, err := dnslib.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
			if err != nil {
				s.logger.Error("Failed to create DNS record for", q.Name, err)
				m.Rcode = dnslib.RcodeServerFailure
				continue
			}
			m.Answer = append(m.Answer, rr)
			s.logger.Info("Local resolution:", q.Name, "->", ip)
			continue
		}

		// Forward unknown query to upstream
		upstream := "1.1.1.1:53"
		s.logger.Info("Forwarding query for", q.Name, "to upstream", upstream)
		resp, err := dnslib.Exchange(r, upstream)
		if err != nil {
			s.logger.Error("Upstream query failed for", q.Name, err)
			m.Rcode = dnslib.RcodeServerFailure
			continue
		}
		m.Answer = append(m.Answer, resp.Answer...)
		s.logger.Info("Upstream response for", q.Name, "- Answers:", len(resp.Answer))
	}

	// Log the response being sent
	s.logger.Info("Sending response to", clientAddr, "- ID:", m.Id, "Answers:", len(m.Answer), "Rcode:", dnslib.RcodeToString[m.Rcode])

	// Log each answer record
	for _, ans := range m.Answer {
		s.logger.Info("Response record:", ans.String())
	}

	_ = w.WriteMsg(m)
}

func (s *Server) Start() error {
	s.dnsServer = &dnslib.Server{Addr: s.cfg.Listen, Net: "udp"}
	dnslib.HandleFunc(".", s.handleDNS)

	s.logger.Info("Starting DNS server on", s.cfg.Listen)
	return s.dnsServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.dnsServer != nil {
		s.logger.Info("Shutting down DNS server...")
		return s.dnsServer.ShutdownContext(ctx)
	}
	return nil
}
