package main

import (
	"log"
	"time"

	"github.com/MuggleWei/cascade"
	"github.com/gorilla/websocket"
)

func connectTimerServ(hub *cascade.Hub, addr string) {
	for {
		c, _, err := websocket.DefaultDialer.Dial(addr, nil)
		if err != nil {
			log.Printf("[Error] failed dial to %v: %v", addr, err.Error())
			time.Sleep(time.Second * 3)
			continue
		}

		server := cascade.NewServer(hub, c, "timerserv")

		server.CallbackOnRead = func(message []byte) {
			server.Hub.MessageChannel <- &cascade.HubMessage{Peer: server, Message: message}
		}

		hub.ServerRegister <- server

		go server.WritePump()
		server.ReadPump(1024)
	}
}
