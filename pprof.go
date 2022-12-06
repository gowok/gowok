package gowok

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func StartPProf() {
	if err := http.ListenAndServe(":6060", nil); err != nil {
		log.Printf("PProf is not running! Reason: %v", err)
	}
}
