package internal

import (
	"context"
	"fmt"
	"strings"
	"waguri-centralized-control/packages/go-utils/config"
	"waguri-centralized-control/packages/go-utils/telemetry"

	dnslib "github.com/miekg/dns"
)

type Server struct {
	cfg       *config.Config
	logger    *telemetry.Logger
	dnsServer *dnslib.Server
}

func NewServer(cfg *config.Config, logger *telemetry.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
	}
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

		// Lookup in config
		ip, ok := s.cfg.Domains[name]
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
