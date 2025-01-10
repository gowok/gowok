package grpc

import (
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/singleton"
	"google.golang.org/grpc"
)

var plugin = "grpc"
var server = singleton.New(func() *grpc.Server {
	return nil
})

func Configure(config *config.Grpc) {
	server(grpc.NewServer())
}

func Server() *grpc.Server {
	return *server()
}
