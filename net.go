package gowok

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
)

var netListen = net.Listen
var netLogFatalln = log.Fatalln

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

	listen, err := netListen(Config.Net.Type, Config.Net.Address)
	if err != nil {
		netLogFatalln("net: failed to start: " + err.Error())
		return
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
	p.Listener = nil
}

func (p *_net) HandleFunc(handler func(net.Conn)) {
	p.handler = handler
}

func (p *_net) Shutdown() {
	if p.Listener != nil {
		_ = p.Close()
	}
}
