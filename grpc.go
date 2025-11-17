package gowok

import (
	"google.golang.org/grpc"
)

type _grpc struct {
	*grpc.Server
}

var GRPC = _grpc{
	Server: grpc.NewServer(),
}
