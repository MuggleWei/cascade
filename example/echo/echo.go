package main

import (
	"log"
	"net/http"
	"os"

	"github.com/MuggleWei/cascade"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 20,
	WriteBufferSize: 1024 * 20,
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)
}

func main() {
	hub := cascade.NewHub()
	go hub.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

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

func serveWs(hub *cascade.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Error] %v\n", err)
		return
	}

	client := cascade.NewPeer(hub, conn)
	client.CallbackOnRead = func(message []byte) {
		log.Printf("[Info] (%v) recv message and echo: %v\n", client.Conn.RemoteAddr().String(), string(message))
		client.SendChannel <- message
	}
	client.Hub.PeerRegister <- client

	go client.WritePump()
	go client.ReadPump(1024)
}
