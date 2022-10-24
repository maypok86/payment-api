package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	defaultHost              string        = "localhost"
	defaultPort              string        = "8080"
	defaultMaxHeaderBytes    int           = 1 << 20 // 1MB
	defaultReadTimeout       time.Duration = 5 * time.Second
	defaultReadHeaderTimeout time.Duration = time.Minute
	defaultWriteTimeout      time.Duration = 5 * time.Second
)

type Server struct {
	host              string
	port              string
	maxHeaderBytes    int
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	httpServer        *http.Server
}

func New(httpHandler http.Handler, opts ...Option) *Server {
	server := &Server{
		host:              defaultHost,
		port:              defaultPort,
		maxHeaderBytes:    defaultMaxHeaderBytes,
		readTimeout:       defaultReadTimeout,
		readHeaderTimeout: defaultReadHeaderTimeout,
		writeTimeout:      defaultWriteTimeout,
	}

	for _, opt := range opts {
		opt(server)
	}

	server.httpServer = &http.Server{
		Addr:              net.JoinHostPort(server.host, server.port),
		Handler:           httpHandler,
		MaxHeaderBytes:    server.maxHeaderBytes,
		ReadTimeout:       server.readTimeout,
		ReadHeaderTimeout: server.readHeaderTimeout,
		WriteTimeout:      server.writeTimeout,
	}

	return server
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("start http server: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context, shutdownTimeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	return nil
}
