package gowok

import (
	"bytes"
	"errors"
	"log/slog"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-must/must"
)

func TestNet(t *testing.T) {
	t.Run("positive/HandleFunc", func(t *testing.T) {
		handlerCalled := false
		Net.HandleFunc(func(conn net.Conn) {
			handlerCalled = true
		})
		Net.handler(nil)
		must.True(t, handlerCalled)
	})

	t.Run("positive/configure and shutdown", func(t *testing.T) {
		Config.Net.Type = "tcp"
		Config.Net.Address = "127.0.0.1:0"

		done := make(chan struct{})
		Net.HandleFunc(func(conn net.Conn) {
			_ = conn.Close()
			done <- struct{}{}
		})

		Net.configure()
		go Net.run()

		for i := 0; i < 100; i++ {
			if Net.Listener != nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		must.NotNil(t, Net.Listener)
		addr := Net.Addr().String()

		conn, err := net.Dial("tcp", addr)
		must.Nil(t, err)
		_ = conn.Close()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("handler was not called")
		}

		Net.Shutdown()
		for i := 0; i < 100; i++ {
			if Net.Listener == nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		must.Nil(t, Net.Listener)
	})

	t.Run("positive/unix socket", func(t *testing.T) {
		socketPath := "/tmp/gowok_test.sock"
		_ = os.Remove(socketPath)
		Config.Net.Type = "unix"
		Config.Net.Address = socketPath

		go Net.configure()

		for i := 0; i < 100; i++ {
			if Net.Listener != nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		must.NotNil(t, Net.Listener)
		if Net.Listener == nil {
			return
		}
		must.Equal(t, "unix", Net.Listener.Addr().Network())
		must.Equal(t, socketPath, Net.Listener.Addr().String())

		Net.Shutdown()
		for i := 0; i < 100; i++ {
			if Net.Listener == nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	})
	t.Run("negative/configure failure", func(t *testing.T) {
		var logBuf bytes.Buffer
		slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, nil)))

		Config.Net.Type = "invalid"
		Config.Net.Address = "invalid"

		Net.configure()
		must.True(t, strings.Contains(logBuf.String(), "level=ERROR"))
	})
	t.Run("negative/Accept failure", func(t *testing.T) {
		oldListen := netListen
		defer func() { netListen = oldListen }()

		netListen = func(network, address string) (net.Listener, error) {
			return &mockListener{
				acceptErr: errors.New("accept error"),
			}, nil
		}

		Config.Net.Type = "tcp"
		Config.Net.Address = "127.0.0.1:0"

		go Net.configure()

		for i := 0; i < 100; i++ {
			if Net.Listener != nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		must.NotNil(t, Net.Listener)
		time.Sleep(100 * time.Millisecond)
		Net.Shutdown()
	})
}

type mockListener struct {
	net.Listener
	acceptErr error
	closed    bool
}

func (m *mockListener) Accept() (net.Conn, error) {
	if m.closed {
		return nil, net.ErrClosed
	}
	if m.acceptErr != nil {
		err := m.acceptErr
		m.acceptErr = nil
		return nil, err
	}
	time.Sleep(10 * time.Millisecond)
	return nil, errors.New("accept error again")
}

func (m *mockListener) Close() error {
	m.closed = true
	return nil
}

func (m *mockListener) Addr() net.Addr {
	return &net.TCPAddr{}
}
