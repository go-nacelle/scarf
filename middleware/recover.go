package middleware

import (
	"context"
	"runtime"

	"github.com/efritz/nacelle"
	"google.golang.org/grpc"

	"github.com/efritz/scarf/logging"
)

type RecoverMiddleware struct {
	Logger           nacelle.Logger     `service:"logger"`
	Decorator        *logging.Decorator `service:"logger-decorator"`
	errorFactory     PanicErrorFactory
	stackBufferSize  int
	logAllGoroutines bool
}

// NewRecovery creates middleware that captures panics from the handler
// and converts them to gRPC error responses. The value of the panic is
// logged at error level.
func NewRecovery(configs ...RecoverConfigFunc) *RecoverMiddleware {
	m := &RecoverMiddleware{
		Logger:           nacelle.NewNilLogger(),
		Decorator:        logging.NewDecorator(),
		errorFactory:     defaultPanicErrorFactory,
		stackBufferSize:  4 << 10,
		logAllGoroutines: false,
	}

	for _, f := range configs {
		f(m)
	}

	return m
}

func (m *RecoverMiddleware) Init() error {
	return nil
}

func (m *RecoverMiddleware) ApplyUnary(f grpc.UnaryHandler, info *grpc.UnaryServerInfo) (grpc.UnaryHandler, error) {
	handler := func(ctx context.Context, req interface{}) (val interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				var (
					stack  = make([]byte, m.stackBufferSize)
					length = runtime.Stack(stack, m.logAllGoroutines)
				)

				m.Decorator.Decorate(ctx, m.Logger).Error(
					"Request handler panicked (%s):\n%s",
					err,
					stack[:length],
				)

				err = m.errorFactory(r)
			}
		}()

		val, err = f(ctx, req)
		return
	}

	return handler, nil
}

func (m *RecoverMiddleware) ApplyStream(f grpc.StreamHandler, info *grpc.StreamServerInfo) (grpc.StreamHandler, error) {
	handler := func(srv interface{}, ss grpc.ServerStream) (err error) {
		defer func() {
			if r := recover(); r != nil {
				var (
					stack  = make([]byte, m.stackBufferSize)
					length = runtime.Stack(stack, m.logAllGoroutines)
				)

				m.Decorator.Decorate(ss.Context(), m.Logger).Error(
					"Request handler panicked (%s):\n%s",
					r,
					stack[:length],
				)

				err = m.errorFactory(r)
			}
		}()

		err = f(srv, ss)
		return
	}

	return handler, nil
}
