package main

import (
	"log"

	"github.com/MuggleWei/cascade"
	"github.com/MuggleWei/cascade/example/common"
)

type CallbackFn func(*cascade.Peer, []byte) error

type ClientInfo struct {
	User    string // user name
	Logined bool   // whether already logined
}

type Manager struct {
	Hub         *cascade.Hub
	ClientInfos map[*cascade.Peer]ClientInfo
	Callbacks   map[string]CallbackFn
}

func NewManager(hub *cascade.Hub) *Manager {
	manager := &Manager{
		Hub:         hub,
		ClientInfos: make(map[*cascade.Peer]ClientInfo),
		Callbacks:   make(map[string]CallbackFn),
	}
	manager.RegisterCallbacks()
	return manager
}

// override hub function
func (this *Manager) OnMessage(message *cascade.HubMessage) {

	user := ""
	if clientInfos, ok := this.ClientInfos[message.Peer]; ok {
		user = clientInfos.User
	}

	op, data_bytes, err := common.ParseStreamData(message.Message)
	if err != nil {
		log.Printf("[Warning] (%v)<%v> failed parse stream data: %v\n",
			message.Peer.Conn.RemoteAddr().String(), user, string(message.Message))
		return
	}

	if callback, ok := this.Callbacks[op]; ok {
		err = callback(message.Peer, data_bytes)
		if err != nil {
		}
	} else {
		log.Printf("[Warning] (%v)<%v> recv message without handle callback: %v\n",
			message.Peer.Conn.RemoteAddr().String(), user, op)
	}
}

func (this *Manager) OnClientActive(client *cascade.Peer) {
	log.Printf("[Info] client active: %v\n", client.Conn.RemoteAddr())
	this.ClientInfos[client] = ClientInfo{
		User:    "",
		Logined: false,
	}
}

func (this *Manager) OnClientInactive(client *cascade.Peer) {
	log.Printf("[Info] client inactive: %v\n", client.Conn.RemoteAddr())
	if _, ok := this.ClientInfos[client]; ok {
		delete(this.ClientInfos, client)
	}
}

// register callbacks
func (this *Manager) RegisterCallbacks() {
	this.Callbacks["login"] = func(client *cascade.Peer, message []byte) error { return this.OnMessageLogin(client, message) }
}

// callback functions
func (this *Manager) OnMessageLogin(client *cascade.Peer, message []byte) error {
	log.Printf("[Info] (%v) req login\n", client.Conn.RemoteAddr().String())
	// check password and add into ClientInfos...
	this.ClientInfos[client] = ClientInfo{
		User:    "xxx",
		Logined: true,
	}
	return nil
}
