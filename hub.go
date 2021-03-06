package cascade

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Slot interface {
	// peer message
	OnActive(*Peer)
	OnInactive(*Peer)
	OnRead(*Peer, []byte)

	// hub message
	OnHubByteMessage(*HubByteMessage)
	OnHubObjectMessage(*HubObjectMessage)
}

type Hub struct {
	Upgrader    *websocket.Upgrader
	MaxReadSize int64

	Peers                map[*Peer]bool         // client's map
	PeerRegister         chan *Peer             // channel that notify peer active
	PeerUnregister       chan *Peer             // channel that notify peer inactive
	ByteMessageChannel   chan *HubByteMessage   // message channel
	ObjectMessageChannel chan *HubObjectMessage // named object channel
	ExitChannel          chan int
	Slot                 Slot
}

type HubByteMessage struct {
	Peer    *Peer
	Message []byte
}

type HubObjectMessage struct {
	Peer       *Peer
	ObjectName string
	ObjectPtr  interface{}
}

func NewHub(slot Slot, upgrader *websocket.Upgrader, maxReadSize int64) *Hub {
	hub := &Hub{
		Upgrader:             upgrader,
		MaxReadSize:          maxReadSize,
		Peers:                make(map[*Peer]bool),
		PeerRegister:         make(chan *Peer),
		PeerUnregister:       make(chan *Peer),
		ByteMessageChannel:   make(chan *HubByteMessage, 100),
		ObjectMessageChannel: make(chan *HubObjectMessage, 100),
		ExitChannel:          make(chan int),
		Slot:                 slot,
	}
	return hub
}

func (this *Hub) Run() {
	for {
		select {
		case objMsg := <-this.ObjectMessageChannel:
			this.Slot.OnHubObjectMessage(objMsg)
		case byteMsg := <-this.ByteMessageChannel:
			this.Slot.OnHubByteMessage(byteMsg)
		case peer := <-this.PeerRegister:
			this.Peers[peer] = true
			this.Slot.OnActive(peer)
		case peer := <-this.PeerUnregister:
			this.Slot.OnInactive(peer)
			if _, ok := this.Peers[peer]; ok {
				delete(this.Peers, peer)
				close(peer.SendChannel)
			}
		case _ = <-this.ExitChannel:
			break
		}
	}
}

type DisconnectCallback func(string, error)

func (this *Hub) ConnectAndRun(addr string, reconn bool, reconnInterval int, reqHeader http.Header, disconnectCallback DisconnectCallback) {
	go this.Run()
	defer this.Stop()

	for {
		conn, _, err := websocket.DefaultDialer.Dial(addr, reqHeader)
		if err != nil {
			// log.Printf("[Error] failed dial to %v: %v", addr, err.Error())
			if disconnectCallback != nil {
				disconnectCallback(addr, err)
			}
			if !reconn {
				break
			}

			if reconnInterval > 0 {
				time.Sleep(time.Second * time.Duration(reconnInterval))
				continue
			}
		}

		peer := NewPeer(this, conn)
		peer.CallbackOnRead = func(message []byte) {
			this.Slot.OnRead(peer, message)
		}
		peer.Hub.PeerRegister <- peer

		go peer.WritePump()
		peer.ReadPump(this.MaxReadSize)

		if !reconn {
			break
		}
	}
}

func (this *Hub) Stop() {
	this.ExitChannel <- 0
}

// It's OK to leave a Go channel open forever and never close it. When the channel
// is no longer used, it will be garbage collected
// see:
//     https://stackoverflow.com/questions/8593645/is-it-ok-to-leave-a-channel-open
//     https://groups.google.com/forum/#!msg/golang-nuts/pZwdYRGxCIk/qpbHxRRPJdUJ
// func (this *Hub) Close() {
// 	close(this.PeerRegister)
// 	close(this.PeerUnregister)
// 	close(this.ByteMessageChannel)
// 	close(this.ObjectMessageChannel)
// 	close(this.ExitChannel)
// }

func (this *Hub) OnAccept(w http.ResponseWriter, r *http.Request) {
	conn, err := this.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Error] %v\n", err)
		return
	}

	peer := NewPeer(this, conn)
	peer.CallbackOnRead = func(message []byte) {
		this.Slot.OnRead(peer, message)
	}
	peer.Header = r.Header
	peer.Hub.PeerRegister <- peer

	go peer.WritePump()
	go peer.ReadPump(this.MaxReadSize)
}
