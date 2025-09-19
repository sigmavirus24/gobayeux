package gobayeuxtest

type ServerOpts interface {
	apply(s *Server)
}

type serverOptFn func(s *Server)

func (opt serverOptFn) apply(s *Server) {
	opt(s)
}

func WithHandshakeError(handshakeError bool) ServerOpts {
	return serverOptFn(func(s *Server) {
		s.handshakeError = handshakeError
	})
}
