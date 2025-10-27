package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gowok/gowok"
)

func Process(ctx context.Context) {
	fmt.Println(ctx.Value("kopi"))
}

type broker struct {
	clients []chan struct{}
}

var b broker

func (b broker) Send() {
	for _, c := range b.clients {
		c <- struct{}{}
	}
}

func main() {
	b = broker{}
	gowok.Router().Get("/events", gowok.HandlerSse(sseHandler))
	gowok.Router().Get("/fire", func(w http.ResponseWriter, r *http.Request) {
		b.Send()
	})
	gowok.Run()
}

func sseHandler(ctx *gowok.WebSseCtx) {
	ctx.Publish([]byte("mantap " + time.Now().String()))
	client := make(chan struct{}, 1)
	b.clients = append(b.clients, client)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Client disconnected")
			return
		case <-client:
			ctx.Emit("asoy", []byte("mantap "+time.Now().String()))
		}
	}
}
