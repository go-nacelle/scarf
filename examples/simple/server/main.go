package main

import (
	"context"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-nacelle/nacelle"
	"github.com/google/uuid"
	"google.golang.org/grpc"

	"github.com/go-nacelle/scarf"
	"github.com/go-nacelle/scarf/examples/simple/proto"
	"github.com/go-nacelle/scarf/middleware"
)

type EndpointSet struct {
	Logger  nacelle.Logger `service:"logger"`
	secrets map[string]string
	mutex   sync.Mutex
}

func NewEndpointSet() *EndpointSet {
	return &EndpointSet{
		secrets: map[string]string{},
	}
}

func (es *EndpointSet) Init(config nacelle.Config, s *grpc.Server) error {
	proto.RegisterSecretServiceServer(s, es)
	return nil
}

func (es *EndpointSet) Middleware() []scarf.Middleware {
	return []scarf.Middleware{
		middleware.NewRequestID(),
		middleware.NewLogging(),
		middleware.NewRecovery(),
	}
}

func (es *EndpointSet) PostSecret(ctx context.Context, req *proto.Secret) (*proto.Id, error) {
	raw, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	name := raw.String()

	es.mutex.Lock()
	es.secrets[name] = req.Secret
	es.mutex.Unlock()

	return &proto.Id{Name: name}, nil
}

func (es *EndpointSet) ReadSecret(ctx context.Context, id *proto.Id) (*proto.Secret, error) {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	secret, ok := es.secrets[id.Name]
	if !ok {
		return nil, status.Error(codes.NotFound, "secret not found")
	}

	delete(es.secrets, id.Name)
	return &proto.Secret{Secret: secret}, nil
}

func main() {
	// TODO - expose additional config registration functions
	scarf.BootAndExit("app", NewEndpointSet())
}
