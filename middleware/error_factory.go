package middleware

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorFactory func(error) error

func defaultErrorFactory(val error) error {
	return status.Error(codes.Internal, val.Error())
}

type PanicErrorFactory func(interface{}) error

func defaultPanicErrorFactory(val interface{}) error {
	return status.Error(codes.Internal, fmt.Sprintf("%#v", val))
}
