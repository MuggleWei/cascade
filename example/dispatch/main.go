package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MuggleWei/cascade"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 20,
	WriteBufferSize: 1024 * 20,
}

func init() {
	//	log.SetOutput(&lumberjack.Logger{
	//		Filename:   "./log/gate.log",
	//		MaxSize:    100,   // MB
	//		MaxBackups: 30,    // old files
	//		MaxAge:     30,    // day
	//		Compress:   false, // disabled by default
	//	})
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)
}

func main() {
	hub := cascade.NewHub()
	manager := NewManager(hub)
	hub.CallbackOnMsg = func(message *cascade.HubMessage) {
		log.Println("on message")
		manager.OnMessage(message)
	}
	hub.CallbackOnClientActive = func(client *cascade.Peer) { manager.OnClientActive(client) }
	hub.CallbackOnClientInactive = func(client *cascade.Peer) { manager.OnClientInactive(client) }
	go hub.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
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

func serveWs(hub *cascade.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Error] %v\n", err)
		return
	}

	client := cascade.NewClient(hub, conn)
	client.CallbackOnRead = func(message []byte) { hub.MessageChannel <- &cascade.HubMessage{Peer: client, Message: message} }
	client.Hub.ClientRegister <- client

	go client.WritePump(time.Second * 5)
	go client.ReadPump(1024)
}
