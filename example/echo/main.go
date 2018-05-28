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
	service := NewEchoService()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		service.Hub.OnAccept(w, r)
	})

	server := &http.Server{
		Addr:    "0.0.0.0:10102",
		Handler: mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("[Fatal] ListenAndServe: %v\n", err)
	}

}
