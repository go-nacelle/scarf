package scarf

import (
	"github.com/efritz/nacelle"
	basegrpc "github.com/efritz/nacelle/base/grpc"
	"google.golang.org/grpc"
)

type (
	serverShim struct {
		Services    nacelle.ServiceContainer `service:"container"`
		endpointSet EndpointSet
		server      *basegrpc.Server
	}
)

func NewServer(endpointSet EndpointSet) nacelle.Process {
	return &serverShim{
		endpointSet: endpointSet,
	}
}

func (s *serverShim) Init(config nacelle.Config) error {
	middleware := s.endpointSet.Middleware()

	for _, m := range middleware {
		if err := s.Services.Inject(m); err != nil {
			return err
		}
	}

	for _, m := range middleware {
		if err := m.Init(); err != nil {
			return err
		}
	}

	s.server = basegrpc.NewServer(
		s.endpointSet,
		basegrpc.WithServerOptions(
			grpc.UnaryInterceptor(makeUnaryInterceptor(middleware)),
			grpc.StreamInterceptor(makeStreamInterceptor(middleware)),
			grpc.StatsHandler(NewStatsHandler(nil)), // TODO - configure
		),
	)

	if err := s.Services.Inject(s.server); err != nil {
		return err
	}

	err := s.server.Init(config)
	return err
}

func (s *serverShim) Start() error {
	return s.server.Start()
}

func (s *serverShim) Stop() error {
	return s.server.Stop()
}
