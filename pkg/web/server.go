package web

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/marcelofabianov/fault"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	router     http.Handler
	addr       string
	tlsConfig  *TLSConfig
}

func NewServer(cfg *Config, logger *slog.Logger, router http.Handler) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)

	server := &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      router,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
			IdleTimeout:  cfg.HTTP.IdleTimeout,
		},
		logger:    logger,
		router:    router,
		addr:      addr,
		tlsConfig: &cfg.HTTP.TLS,
	}

	if cfg.HTTP.TLS.Enabled {
		server.httpServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
			CipherSuites: []uint16{
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			},
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
			},
		}
	}

	return server
}

func (s *Server) Start() error {
	if s.tlsConfig.Enabled {
		s.logger.Info("Starting HTTPS server with TLS 1.2/1.3",
			"addr", s.addr,
			"cert_file", s.tlsConfig.CertFile,
			"key_file", s.tlsConfig.KeyFile,
		)

		if err := s.httpServer.ListenAndServeTLS(s.tlsConfig.CertFile, s.tlsConfig.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fault.Wrap(err, "failed to start HTTPS server", fault.WithCode(fault.Internal))
		}
	} else {
		s.logger.Info("Starting HTTP server", "addr", s.addr)

		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fault.Wrap(err, "failed to start HTTP server", fault.WithCode(fault.Internal))
		}
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server", "addr", s.addr)

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fault.Wrap(err, "failed to shutdown HTTP server", fault.WithCode(fault.Internal))
	}

	s.logger.Info("HTTP server shutdown complete")
	return nil
}

func (s *Server) Addr() string {
	return s.addr
}
