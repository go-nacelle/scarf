package scarf

import (
	"context"

	"google.golang.org/grpc"
)

type (
	// Middleware transforms unary and stream handlers into a decorated handler.
	Middleware interface {
		Init() error

		// ApplyUnary decorates a UnaryHandler.
		ApplyUnary(grpc.UnaryHandler, *grpc.UnaryServerInfo) (grpc.UnaryHandler, error)

		// ApplyStream decorates a UnaryHandler.
		ApplyStream(grpc.StreamHandler, *grpc.StreamServerInfo) (grpc.StreamHandler, error)
	}
)

func makeUnaryInterceptor(middleware []Middleware) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, f grpc.UnaryHandler) (val interface{}, err error) {
		// TODO - backwards?
		for i := len(middleware) - 1; i >= 0; i-- {
			f, err = middleware[i].ApplyUnary(f, info)
			if err != nil {
				return
			}
		}

		val, err = f(ctx, req)
		return
	}
}

func makeStreamInterceptor(middleware []Middleware) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, f grpc.StreamHandler) (err error) {
		// TODo - backwards?
		for i := len(middleware) - 1; i >= 0; i-- {
			f, err = middleware[i].ApplyStream(f, info)
			if err != nil {
				return
			}
		}

		err = f(srv, ss)
		return
	}
}
