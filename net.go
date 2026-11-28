package gowok

import (
	"errors"
	"log/slog"
	"net"
	"os"
	"sync"
)

var netListen = net.Listen

type _net struct {
	sync.RWMutex
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

	p.Lock()
	p.Listener = listen
	p.Unlock()
}

func (p *_net) getListener() net.Listener {
	p.RLock()
	defer p.RUnlock()
	return p.Listener
}

func (p *_net) HandleFunc(handler func(net.Conn)) {
	p.Lock()
	p.handler = handler
	p.Unlock()
}

func (p *_net) Shutdown() {
	p.Lock()
	l := p.Listener
	p.Listener = nil
	p.Unlock()

	if l != nil {
		_ = l.Close()
	}
}

func (p *_net) run() {
	slog.Info("starting net")

	p.RLock()
	listener := p.Listener
	handler := p.handler
	p.RUnlock()

	if listener == nil {
		slog.Error("net: failed to start", "error", "bad configurations")
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}

			slog.Error("net: failed to process", "error", err)
			continue
		}
		go handler(conn)
	}

	p.Lock()
	if p.Listener == listener {
		p.Listener = nil
	}
	p.Unlock()
}
