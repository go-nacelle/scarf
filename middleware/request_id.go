package middleware

import (
	"context"

	"github.com/efritz/nacelle"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/efritz/scarf/logging"
)

type (
	RequestIDMiddleware struct {
		Logger             nacelle.Logger     `service:"logger"`
		Decorator          *logging.Decorator `service:"logger-decorator"`
		requestIDGenerator RequestIDGenerator
		errorFactory       ErrorFactory
	}

	RequestIDGenerator func() (string, error)
	tokenRequestID     string
)

// TokenRequestID is the unique token to which the current request's unique
// ID is written to the request context.
var TokenRequestID = tokenRequestID("scarf.middleware.request_id")

// GetRequestID retrieves the current request's unique ID. If no request ID
// is registered with this context, the empty string is returned.
func GetRequestID(ctx context.Context) string {
	if val, ok := ctx.Value(TokenRequestID).(string); ok {
		return val
	}

	return ""
}

// NewRequestID creates middleware that generates a unique ID for the request.
// The request ID is added to the context and is made available by the
// GetRequestID function.
func NewRequestID(configs ...RequestIDConfigFunc) *RequestIDMiddleware {
	m := &RequestIDMiddleware{
		Logger:             nacelle.NewNilLogger(),
		Decorator:          logging.NewDecorator(),
		requestIDGenerator: defaultRequestIDGenerator,
		errorFactory:       defaultErrorFactory,
	}

	for _, f := range configs {
		f(m)
	}

	return m
}

func (m *RequestIDMiddleware) Init() error {
	m.Decorator.Register(func(ctx context.Context, fields nacelle.LogFields) {
		if requestID := GetRequestID(ctx); requestID != "" {
			fields["request_id"] = requestID
		}
	})

	return nil
}

func (m *RequestIDMiddleware) ApplyUnary(f grpc.UnaryHandler, info *grpc.UnaryServerInfo) (grpc.UnaryHandler, error) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		requestID, err := m.getIDFromRequest(ctx)
		if err != nil {
			m.Decorator.Decorate(ctx, m.Logger).Error(
				"Failed to generate request ID (%s)",
				err.Error(),
			)

			return nil, m.errorFactory(err)
		}

		return f(context.WithValue(ctx, TokenRequestID, requestID), req)
	}

	return handler, nil
}

func (m *RequestIDMiddleware) ApplyStream(f grpc.StreamHandler, info *grpc.StreamServerInfo) (grpc.StreamHandler, error) {
	handler := func(srv interface{}, ss grpc.ServerStream) error {
		requestID, err := m.getIDFromRequest(ss.Context())
		if err != nil {
			m.Decorator.Decorate(ss.Context(), m.Logger).Error(
				"Failed to generate request ID (%s)",
				err.Error(),
			)

			return m.errorFactory(err)
		}

		return f(srv, WrapServerStream(ss, TokenRequestID, requestID))
	}

	return handler, nil
}

func (m *RequestIDMiddleware) getIDFromRequest(ctx context.Context) (string, error) {
	if metadata, ok := metadata.FromIncomingContext(ctx); ok {
		if requestID, ok := metadata["x-request-id"]; ok {
			return requestID[0], nil
		}
	}

	return m.requestIDGenerator()
}

func defaultRequestIDGenerator() (string, error) {
	raw, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return raw.String(), nil
}
