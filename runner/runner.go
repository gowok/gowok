package runner

import (
	"os"
	"os/signal"
	"syscall"
)

func GracefulStop(callback func()) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	func() {
		<-gracefulStop

		callback()

		os.Exit(0)
	}()
}
