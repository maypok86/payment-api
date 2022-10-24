package server

import "time"

type Option func(*Server)

func WithHost(host string) Option {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port string) Option {
	return func(s *Server) {
		s.port = port
	}
}

func WithMaxHeaderBytes(maxHeaderBytes int) Option {
	return func(s *Server) {
		s.maxHeaderBytes = maxHeaderBytes * (1 << 20)
	}
}

func WithReadTimeout(readTimeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = readTimeout
	}
}

func WithReadHeaderTimeout(readHeaderTimeout time.Duration) Option {
	return func(s *Server) {
		s.readHeaderTimeout = readHeaderTimeout
	}
}

func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = writeTimeout
	}
}
