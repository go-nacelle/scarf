package scarf

import (
	"github.com/efritz/nacelle"
	"google.golang.org/grpc"
)

type EndpointSet interface {
	Init(nacelle.Config, *grpc.Server) error
	Middleware() []Middleware
}
