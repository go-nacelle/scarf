package middleware

import (
	"context"
	"time"

	"github.com/derision-test/glock"
	"github.com/go-nacelle/nacelle"
	"google.golang.org/grpc"

	"github.com/go-nacelle/scarf/logging"
)

type LoggingMiddleware struct {
	Logger    nacelle.Logger     `service:"logger"`
	Decorator *logging.Decorator `service:"logger-decorator"`
	clock     glock.Clock
}

// NewLogging creates middleware that logs incoming requests and
// the status of the response after the request is handled.
func NewLogging(configs ...LoggingConfigFunc) *LoggingMiddleware {
	m := &LoggingMiddleware{
		Logger:    nacelle.NewNilLogger(),
		Decorator: logging.NewDecorator(),
		clock:     glock.NewRealClock(),
	}

	for _, f := range configs {
		f(m)
	}

	return m
}

func (m *LoggingMiddleware) Init() error {
	return nil
}

func (m *LoggingMiddleware) ApplyUnary(f grpc.UnaryHandler, info *grpc.UnaryServerInfo) (grpc.UnaryHandler, error) {
	handler := func(ctx context.Context, req interface{}) (val interface{}, err error) {
		m.logRPC(ctx, info.FullMethod, false, false, func() error {
			val, err = f(ctx, req)
			return err
		})

		return
	}

	return handler, nil
}

func (m *LoggingMiddleware) ApplyStream(f grpc.StreamHandler, info *grpc.StreamServerInfo) (grpc.StreamHandler, error) {
	handler := func(srv interface{}, ss grpc.ServerStream) error {
		return m.logRPC(ss.Context(), info.FullMethod, info.IsClientStream, info.IsServerStream, func() error {
			return f(srv, ss)
		})
	}

	return handler, nil
}

func (m *LoggingMiddleware) logRPC(
	ctx context.Context,
	fullMethod string,
	isClientStream bool,
	isServerStream bool,
	f func() error,
) error {
	logger := m.Decorator.Decorate(ctx, m.Logger).WithFields(nacelle.LogFields{
		"full_method":   fullMethod,
		"client_stream": isClientStream,
		"server_stream": isServerStream,
	})

	logger.Info(
		"Handling rpc request %s",
		fullMethod,
	)

	start := m.clock.Now()
	err := f()
	duration := int(m.clock.Now().Sub(start) / time.Millisecond)

	if err != nil {
		logger.Error(
			"Request to %s returned an error (%s)",
			fullMethod,
			err.Error(),
		)

		return err
	}

	logger.Info(
		"Handled rpc request %s in %dms",
		fullMethod,
		duration,
	)

	return nil
}
