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
	setHubCallback(hub)
	go hub.Run()

	go connectTimerServ(hub, "ws://127.0.0.1:10000/timer")

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

	client := cascade.NewClient(hub, conn)
	client.Hub.ClientRegister <- client

	go client.WritePump()
	go client.ReadPump(1024)
}

func setHubCallback(hub *cascade.Hub) {
	hub.CallbackOnMsg = func(message *cascade.HubMessage) {
		for client := range hub.Clients {
			select {
			case client.SendChannel <- message.Message:
			default:
				log.Printf("[Warning] SendChannel full\n")
				//		close(client.SendChannel)
				//		delete(hub.Clients, client)
			}
		}
	}
}
