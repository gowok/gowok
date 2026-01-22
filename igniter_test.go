package gowok

import (
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/golang-must/must"
	"github.com/gowok/gowok/config"
)

func resetIgniter() {
	_project(nil)
	Config = &config.Config{}
}

func TestIgniter_Configure(t *testing.T) {
	oldLogFatalln := igniterLogFatalln
	defer func() {
		igniterLogFatalln = oldLogFatalln
		resetIgniter()
	}()

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
		fatalCalled := false
		igniterLogFatalln = func(v ...any) {
			fatalCalled = true
		}
		configure("non-existent-file.yaml")
		must.True(t, fatalCalled)
	})
}

func TestIgniter_Run(t *testing.T) {
	oldListenAndServe := webListenAndServe
	oldNetListen := netListen
	oldWebLogFatalln := webLogFatalln
	oldNetLogFatalln := netLogFatalln
	defer func() {
		webListenAndServe = oldListenAndServe
		netListen = oldNetListen
		webLogFatalln = oldWebLogFatalln
		netLogFatalln = oldNetLogFatalln
		resetIgniter()
	}()

	webListenAndServe = func(s *http.Server) error { return http.ErrServerClosed }
	netListen = func(network, address string) (net.Listener, error) { return nil, errors.New("mock error") }
	webLogFatalln = func(v ...any) {}
	netLogFatalln = func(v ...any) {}

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

		// Wait for goroutines to finish since we mocked failures
		time.Sleep(10 * time.Millisecond)
	})
}

func TestIgniter_Stop(t *testing.T) {
	oldLogFatalln := igniterLogFatalln
	defer func() {
		igniterLogFatalln = oldLogFatalln
		resetIgniter()
	}()

	t.Run("positive/stop web and net", func(t *testing.T) {
		resetIgniter()
		c := config.Config{
			Web: config.Web{Enabled: true},
			Net: config.Net{Enabled: true},
		}
		configure(c)

		// Mock Web.Server to avoid nil pointer if needed, though it's initialized in Web global
		Web.Server = &http.Server{}

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
		// Just making sure it doesn't panic
		Shutdown()
	})
}

func TestIgniter_RunPublic(t *testing.T) {
	defer func() { resetIgniter() }()

	t.Run("positive/run", func(t *testing.T) {
		resetIgniter()
		oldWebLogFatalln := webLogFatalln
		oldNetLogFatalln := netLogFatalln
		webLogFatalln = func(v ...any) {}
		netLogFatalln = func(v ...any) {}
		defer func() {
			webLogFatalln = oldWebLogFatalln
			netLogFatalln = oldNetLogFatalln
		}()

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
