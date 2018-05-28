package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)
}

func main() {
	listenAddr := flag.String("addr", "127.0.0.1:10000", "listen address")
	flag.Parse()

	service := NewTimerService()

	runTicker(service.Hub)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		service.Hub.OnAccept(w, r)
	})

	server := &http.Server{
		Addr:    *listenAddr,
		Handler: mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("[Fatal] ListenAndServe: %v\n", err)
	}
}
