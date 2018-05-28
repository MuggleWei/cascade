package main

import (
	"log"

	"github.com/MuggleWei/cascade"
	"github.com/gorilla/websocket"
)

type EchoService struct {
	Hub *cascade.Hub
}

func NewEchoService() *EchoService {
	service := &EchoService{}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024 * 20,
		WriteBufferSize: 1024 * 20,
	}

	hub := cascade.NewHub(service, &upgrader, 10240)
	go hub.Run()

	service.Hub = hub
	return service
}

func (this *EchoService) OnActive(peer *cascade.Peer) {
	log.Printf("OnActive: %v\n", peer.Conn.RemoteAddr().String())
}

func (this *EchoService) OnInactive(peer *cascade.Peer) {
	log.Printf("OnInactive: %v\n", peer.Conn.RemoteAddr().String())
}

func (this *EchoService) OnRead(peer *cascade.Peer, message []byte) {
	log.Printf("On message: %v\n", string(message))
	peer.SendChannel <- message
}

func (this *EchoService) OnHubByteMessage(msg *cascade.HubByteMessage) {
}

func (this *EchoService) OnHubObjectMessage(*cascade.HubObjectMessage) {
}
