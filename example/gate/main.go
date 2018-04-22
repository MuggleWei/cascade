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

	timerHub := cascade.NewHub()
	setTimerServHubCallback(timerHub, hub)
	go connectTimerServ(timerHub, "ws://127.0.0.1:10000/timer")

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
	client.Hub.PeerRegister <- client

	go client.WritePump()
	go client.ReadPump(1024)
}

func setHubCallback(hub *cascade.Hub) {
	hub.CallbackOnActive = func(client *cascade.Peer) {
		log.Printf("client active: %v\n", client.Conn.RemoteAddr())
	}
	hub.CallbackOnInactive = func(client *cascade.Peer) {
		log.Printf("client inactive: %v\n", client.Conn.RemoteAddr())
	}
	hub.CallbackOnMsg = func(message *cascade.HubMessage) {
		for client := range hub.Peers {
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

func setTimerServHubCallback(timerHub, hub *cascade.Hub) {
	timerHub.CallbackOnActive = func(timerServ *cascade.Peer) {
		log.Printf("connected to timer server: %v\n", timerServ.Conn.RemoteAddr())
	}
	timerHub.CallbackOnInactive = func(timerServ *cascade.Peer) {
		log.Printf("disconnected to timer server: %v\n", timerServ.Conn.RemoteAddr())
	}
	timerHub.CallbackOnMsg = func(message *cascade.HubMessage) {
		hub.MessageChannel <- message
	}
}
