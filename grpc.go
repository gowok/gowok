package gowok

import (
	"log"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type _grpc struct {
	*grpc.Server
}

var GRPC = _grpc{
	Server: grpc.NewServer(),
}

func (p *_grpc) configure() {
	slog.Info("starting GRPC", "host", Config.Grpc.Host)
	listen, err := net.Listen("tcp", Config.Grpc.Host)
	if err != nil {
		log.Fatalln("grpc: failed to start: " + err.Error())
	}

	err = GRPC.Serve(listen)
	if err != nil {
		log.Fatalln("grpc: failed to start: " + err.Error())
	}
}
