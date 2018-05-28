package main

import (
	"encoding/json"
	"log"

	"github.com/MuggleWei/cascade"
	"github.com/MuggleWei/cascade/example"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

type CallbackFn func(*cascade.Peer, *example.CommonMessage)

type DispatchService struct {
	Hub       *cascade.Hub
	Callbacks map[string]CallbackFn
}

func NewDispatchService() *DispatchService {
	service := &DispatchService{
		Hub:       nil,
		Callbacks: make(map[string]CallbackFn),
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024 * 20,
		WriteBufferSize: 1024 * 20,
	}

	hub := cascade.NewHub(service, &upgrader, 10240)
	go hub.Run()

	service.Hub = hub
	service.RegisterCallbacks()

	return service
}

// Slot callbacks
func (this *DispatchService) OnActive(peer *cascade.Peer) {
	log.Printf("OnActive: %v\n", peer.Conn.RemoteAddr().String())
}

func (this *DispatchService) OnInactive(peer *cascade.Peer) {
	log.Printf("OnInactive: %v\n", peer.Conn.RemoteAddr().String())
}

func (this *DispatchService) OnRead(peer *cascade.Peer, message []byte) {
	var msg example.CommonMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		panic(err)
	}

	callbackFn, ok := this.Callbacks[msg.Op]
	if ok {
		callbackFn(peer, &msg)
	} else {
		log.Printf("failed get message callback: %+v\n", msg.Op)
	}
}

func (this *DispatchService) OnHubByteMessage(msg *cascade.HubByteMessage) {
}

func (this *DispatchService) OnHubObjectMessage(*cascade.HubObjectMessage) {
}

// dispatch
func (this *DispatchService) RegisterCallbacks() {
	this.Callbacks["login"] = this.OnLogin
	this.Callbacks["logout"] = this.OnLogout
	this.Callbacks["greet"] = this.OnGreet
}

// callbacks
func (this *DispatchService) OnLogin(peer *cascade.Peer, message *example.CommonMessage) {
	var loginMsg example.LoginMessage
	err := mapstructure.Decode(message.Data, &loginMsg)
	if err != nil {
		panic(err)
	}
	log.Printf("OnLogin: %v - %+v\n", peer.Conn.RemoteAddr().String(), loginMsg)
}

func (this *DispatchService) OnLogout(peer *cascade.Peer, message *example.CommonMessage) {
	log.Println("OnLogout")
}

func (this *DispatchService) OnGreet(peer *cascade.Peer, message *example.CommonMessage) {
	log.Println("OnGreet")
}
