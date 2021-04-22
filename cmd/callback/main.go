package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/evt/callback/config"

	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// config
	cfg := config.Get()

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	go func() { _ = http.ListenAndServe(cfg.HTTPAddr, nil) }()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

	fmt.Println("closing")

	return nil
}
