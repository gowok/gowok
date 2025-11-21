package gowok

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
)

type _net struct {
	net.Listener
	handler func(net.Conn)
}

var Net = &_net{
	handler: func(l net.Conn) {},
}

func (p *_net) configure() {
	slog.Info("starting net")

	if Config.Net.Type == "unix" {
		_ = os.Remove(Config.Net.Address)
	}

	listen, err := net.Listen(Config.Net.Type, Config.Net.Address)
	if err != nil {
		log.Fatalln("net: failed to start: " + err.Error())
	}

	p.Listener = listen
	for {
		conn, err := listen.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}

			fmt.Println("net: failed to process: " + err.Error())
			continue
		}
		go p.handler(conn)
	}
}

func (p *_net) HandleFunc(handler func(net.Conn)) {
	p.handler = handler
}

func (p *_net) Shutdown() {
	_ = p.Listener.Close()
}
