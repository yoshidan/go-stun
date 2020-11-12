package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yoshidan/go-stun/stun"
)

func main() {

	server := stun.NewServer(context.Background(), ":3478")

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)
		<-sigint
		server.Shutdown()
	}()

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	log.Printf("Shutdown server.")
}
