package gowok

import (
	"errors"
	"log/slog"
	"net"
	"os"
)

var netListen = net.Listen

type _net struct {
	net.Listener
	handler func(net.Conn)
}

var Net = &_net{
	handler: func(l net.Conn) {},
}

func (p *_net) configure() {
	if Config.Net.Type == "unix" {
		_ = os.Remove(Config.Net.Address)
	}

	listen, err := netListen(Config.Net.Type, Config.Net.Address)
	if err != nil {
		slog.Error("net: failed to start", "error", err)
		return
	}

	p.Listener = listen
}

func (p *_net) HandleFunc(handler func(net.Conn)) {
	p.handler = handler
}

func (p *_net) Shutdown() {
	if p.Listener != nil {
		_ = p.Close()
	}
}

func (p *_net) run() {
	slog.Info("starting net")

	if p.Listener == nil {
		slog.Error("net: failed to start", "error", "bad configurations")
		return
	}

	for {
		conn, err := p.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}

			slog.Error("net: failed to process", "error", err)
			continue
		}
		go p.handler(conn)
	}
	p.Listener = nil
}
