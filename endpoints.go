package scarf

import (
	"github.com/go-nacelle/nacelle"
	"google.golang.org/grpc"
)

type EndpointSet interface {
	Init(nacelle.Config, *grpc.Server) error
	Middleware() []Middleware
}
