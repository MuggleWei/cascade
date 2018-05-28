package main

import (
	"log"
	"net/http"
	"os"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)
}

func main() {
	gate := NewGateService()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		gate.Hub.OnAccept(w, r)
	})

	timerService := NewTimerService()
	timerService.GateHub = gate.Hub
	go timerService.Hub.ConnectAndRun("ws://127.0.0.1:10000/ws", true, 3)

	server := &http.Server{
		Addr:    "0.0.0.0:10102",
		Handler: mux,
	}
	// err := server.ListenAndServeTLS("ca/server.crt", "ca/server.key")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("[Fatal] ListenAndServe: %v\n", err)
	}
}
