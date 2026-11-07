package gowok

import (
	"bytes"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-must/must"
	"github.com/gowok/gowok/config"
	"github.com/ngamux/ngamux"
)

func resetIgniter() {
	_project(nil)
	Config = &config.Config{}
	Web = &_web{
		HttpServeMux: ngamux.NewHttpServeMux(),
		Server:       &http.Server{},
		Handler:      &_webHandler{},
		Response:     &_webResponse{},
		Request:      &_webRequest{},
		Resource:     &_webResource{},
	}
}

func TestIgniter_Configure(t *testing.T) {
	defer resetIgniter()

	t.Run("positive/basic configuration", func(t *testing.T) {
		resetIgniter()
		p := configure()
		must.NotNil(t, p)
	})

	t.Run("positive/from config struct", func(t *testing.T) {
		resetIgniter()
		c := config.Config{Key: "test-key"}
		p := configure(c)
		must.Equal(t, "test-key", Config.Key)
		must.NotNil(t, p)
	})

	t.Run("negative/invalid file path", func(t *testing.T) {
		resetIgniter()
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		configure("non-existent-file.yaml")
		w.Close()
		os.Stdout = oldStdout
		var logBuf bytes.Buffer
		logBuf.ReadFrom(r)
		must.True(t, strings.Contains(logBuf.String(), `"level":"ERROR"`))
	})
}

func TestIgniter_Run(t *testing.T) {
	oldListenAndServe := webListenAndServe
	oldNetListen := netListen
	defer func() {
		webListenAndServe = oldListenAndServe
		netListen = oldNetListen
		resetIgniter()
	}()

	webListenAndServe = func(s *http.Server) error { return http.ErrServerClosed }
	netListen = func(network, address string) (net.Listener, error) { return nil, errors.New("mock error") }

	t.Run("positive/run web and net", func(t *testing.T) {
		resetIgniter()
		c := config.Config{
			Web: config.Web{Enabled: true},
			Net: config.Net{Enabled: true},
		}
		p := configure(c)
		Config.Forever = false

		startingCalled := false
		startedCalled := false
		Hooks.SetOnStarting(func() { startingCalled = true })
		Hooks.SetOnStarted(func() { startedCalled = true })

		p.run()

		must.True(t, startingCalled)
		must.True(t, startedCalled)

		time.Sleep(10 * time.Millisecond)
	})
}

func TestIgniter_Stop(t *testing.T) {
	oldNetListen := netListen
	defer func() {
		netListen = oldNetListen
		resetIgniter()
	}()

	netListen = func(network, address string) (net.Listener, error) { return nil, errors.New("mock error") }

	t.Run("positive/stop web and net", func(t *testing.T) {
		resetIgniter()
		c := config.Config{
			Web: config.Web{Enabled: true},
			Net: config.Net{Enabled: true},
		}

		configure(c)

		stoppedCalled := false
		Hooks.SetOnStopped(func() { stoppedCalled = true })

		stop()
		must.True(t, stoppedCalled)
	})
}

func TestIgniter_Shutdown(t *testing.T) {
	defer func() { resetIgniter() }()

	t.Run("positive/shutdown", func(t *testing.T) {
		resetIgniter()
		configure()
		Shutdown()
	})
}

func TestIgniter_RunPublic(t *testing.T) {
	defer func() { resetIgniter() }()

	t.Run("positive/run", func(t *testing.T) {
		resetIgniter()

		Config.Forever = false
		Run(config.Config{Key: "test"})
		must.Equal(t, "test", Config.Key)
		time.Sleep(10 * time.Millisecond)
	})
}

func TestIgniter_Configures(t *testing.T) {
	defer func() { resetIgniter() }()

	t.Run("positive/configures", func(t *testing.T) {
		resetIgniter()
		called := false
		p := Configures(func() {
			called = true
		})
		must.NotNil(t, p)
		must.True(t, called)
		must.Equal(t, 1, len(p.configures))
	})
}
