package main

import (
	"log"

	"github.com/MuggleWei/cascade"
)

type TimerService struct {
	Hub     *cascade.Hub
	GateHub *cascade.Hub
}

func NewTimerService() *TimerService {
	service := &TimerService{
		Hub:     nil,
		GateHub: nil,
	}

	hub := cascade.NewHub(service, nil, 10240)
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
	this.GateHub.ByteMessageChannel <- &cascade.HubByteMessage{Peer: nil, Message: message}
}

func (this *TimerService) OnHubByteMessage(msg *cascade.HubByteMessage) {
}

func (this *TimerService) OnHubObjectMessage(*cascade.HubObjectMessage) {
}
