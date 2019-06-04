package main

import (
	"io"
	"sync"

	"github.com/go-nacelle/nacelle"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"google.golang.org/grpc"

	"github.com/go-nacelle/scarf"
	"github.com/go-nacelle/scarf/examples/stream/proto"
	"github.com/go-nacelle/scarf/middleware"
)

type (
	EndpointSet struct {
		Logger      nacelle.Logger `service:"logger"`
		value       int32
		updates     int32
		subscribers map[string]chan *proto.Summary
		mutex       sync.RWMutex
	}
)

func NewEndpointSet() *EndpointSet {
	return &EndpointSet{
		subscribers: map[string]chan *proto.Summary{},
	}
}

func (es *EndpointSet) Init(config nacelle.Config, s *grpc.Server) error {
	proto.RegisterValueServiceServer(s, es)
	return nil
}

func (es *EndpointSet) Middleware() []scarf.Middleware {
	return []scarf.Middleware{
		middleware.NewRequestID(),
		middleware.NewLogging(),
		middleware.NewRecovery(),
	}
}

func (es *EndpointSet) Subscribe(_ *empty.Empty, stream proto.ValueService_SubscribeServer) error {
	name, ch, err := es.subscribe()
	if err != nil {
		return err
	}

	defer es.unsubscribe(name)

	for summary := range ch {
		if err := stream.Send(summary); err != nil {
			return err
		}
	}

	return nil
}

func (es *EndpointSet) QuietUpdate(stream proto.ValueService_QuietUpdateServer) error {
	for {
		update, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return stream.SendAndClose(es.makeSummary())
			}

			return err
		}

		es.update(update.GetDelta())
	}
}

func (es *EndpointSet) NoisyUpdate(stream proto.ValueService_NoisyUpdateServer) error {
	name, ch, err := es.subscribe()
	if err != nil {
		return err
	}

	defer es.unsubscribe(name)

	go func() {
		for summary := range ch {
			stream.Send(summary)
		}
	}()

	for {
		update, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		es.update(update.GetDelta())
	}
}

//
//

func (es *EndpointSet) subscribe() (string, <-chan *proto.Summary, error) {
	raw, err := uuid.NewRandom()
	if err != nil {
		return "", nil, err
	}

	name := raw.String()
	summaries := make(chan *proto.Summary)

	es.mutex.Lock()
	defer es.mutex.Unlock()

	es.subscribers[name] = summaries
	return name, summaries, nil

}

func (es *EndpointSet) unsubscribe(name string) {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	ch := es.subscribers[name]
	delete(es.subscribers, name)
	close(ch)
}

func (es *EndpointSet) update(delta int32) {
	es.bump(delta)
	es.notify(es.makeSummary())
}

func (es *EndpointSet) bump(delta int32) {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	es.updates++
	es.value += delta
}

func (es *EndpointSet) makeSummary() *proto.Summary {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	return &proto.Summary{
		Updates: es.updates,
		Value:   es.value,
	}
}

func (es *EndpointSet) notify(summary *proto.Summary) {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	for _, summaries := range es.subscribers {
		summaries <- summary
	}
}

//
//

func main() {
	scarf.BootAndExit("app", NewEndpointSet())
}
