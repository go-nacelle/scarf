package middleware

import (
	"context"

	"google.golang.org/grpc"
)

type wrappedServerStream struct {
	grpc.ServerStream
	key   interface{}
	value interface{}
}

func WrapServerStream(ss grpc.ServerStream, key, value interface{}) grpc.ServerStream {
	return &wrappedServerStream{
		ServerStream: ss,
		key:          key,
		value:        value,
	}
}

func (s *wrappedServerStream) Context() context.Context {
	return context.WithValue(s.ServerStream.Context(), s.key, s.value)
}
