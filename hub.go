package cascade

import (
	"log"
)

type Hub struct {
	Clients          map[*Peer]bool   // client's map
	ClientRegister   chan *Peer       // channel that notify client active
	ClientUnregister chan *Peer       // channel that notify client inactive
	Servers          map[string]*Peer // server's map
	ServerRegister   chan *Peer       // channel that notify server active
	ServerUnregister chan *Peer       // channel that notify server inactive
	MessageChannel   chan *HubMessage // message channel

	CallbackOnClientActive   func(*Peer)       // on client active
	CallbackOnClientInactive func(*Peer)       // on client inactive
	CallbackOnServerActive   func(*Peer)       // on server active
	CallbackOnServerInactive func(*Peer)       // on server inactive
	CallbackOnMsg            func(*HubMessage) // on message
}

type HubMessage struct {
	Peer    *Peer
	Message []byte
}

func NewHub() *Hub {
	return &Hub{
		Clients:                  make(map[*Peer]bool),
		ClientRegister:           make(chan *Peer),
		ClientUnregister:         make(chan *Peer),
		Servers:                  make(map[string]*Peer),
		ServerRegister:           make(chan *Peer),
		ServerUnregister:         make(chan *Peer),
		MessageChannel:           make(chan *HubMessage),
		CallbackOnClientActive:   nil,
		CallbackOnClientInactive: nil,
		CallbackOnServerActive:   nil,
		CallbackOnServerInactive: nil,
		CallbackOnMsg:            nil,
	}
}

func (this *Hub) Run() {
	for {
		select {
		case client := <-this.ClientRegister:
			log.Printf("[Info] client (%v) active\n", client.Conn.RemoteAddr())
			this.Clients[client] = true
			if this.CallbackOnClientActive != nil {
				this.CallbackOnClientActive(client)
			}
		case client := <-this.ClientUnregister:
			log.Printf("[Info] client (%v)<%v> inactive\n", client.Conn.RemoteAddr(), client.Name)
			if this.CallbackOnClientInactive != nil {
				this.CallbackOnClientInactive(client)
			}
			if _, ok := this.Clients[client]; ok {
				delete(this.Clients, client)
				close(client.SendChannel)
			}
		case server := <-this.ServerRegister:
			log.Printf("[Info] server (%v)<%v> active\n", server.Conn.RemoteAddr(), server.Name)
			if _, ok := this.Servers[server.Name]; ok {
				close(server.SendChannel)
				log.Printf("[Warning] repeated connect to server <%v>(%v)\n", server.Name, server.Conn.RemoteAddr())
			} else {
				this.Servers[server.Name] = server
				if this.CallbackOnServerActive != nil {
					this.CallbackOnServerActive(server)
				}
			}
		case server := <-this.ServerUnregister:
			log.Printf("[Info] server (%v)<%v> inactive\n", server.Conn.RemoteAddr(), server.Name)
			if this.CallbackOnServerInactive != nil {
				this.CallbackOnServerInactive(server)
			}
			if _, ok := this.Servers[server.Name]; ok {
				delete(this.Servers, server.Name)
				close(server.SendChannel)
			}
		case hubMessage := <-this.MessageChannel:
			if this.CallbackOnMsg != nil {
				this.CallbackOnMsg(hubMessage)
			}
		}
	}
}
