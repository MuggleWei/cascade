package main

import (
	"log"

	"github.com/MuggleWei/cascade"
	"github.com/gorilla/websocket"
)

type TimerService struct {
	Hub *cascade.Hub
}

func NewTimerService() *TimerService {
	service := &TimerService{
		Hub: nil,
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024 * 20,
		WriteBufferSize: 1024 * 20,
	}

	hub := cascade.NewHub(service, &upgrader, 10240)
	go hub.Run()

	service.Hub = hub

	return service
}

// Slot callbacks
func (this *TimerService) OnActive(peer *cascade.Peer) {
	log.Printf("OnActive: %v\n", peer.Conn.RemoteAddr().String())
}

func (this *TimerService) OnInactive(peer *cascade.Peer) {
	log.Printf("OnInactive: %v\n", peer.Conn.RemoteAddr().String())
}

func (this *TimerService) OnRead(peer *cascade.Peer, message []byte) {
}

func (this *TimerService) OnHubByteMessage(msg *cascade.HubByteMessage) {
	for peer := range this.Hub.Peers {
		peer.SendChannel <- msg.Message
	}
}

func (this *TimerService) OnHubObjectMessage(*cascade.HubObjectMessage) {
}
