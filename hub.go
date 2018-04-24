package cascade

type NamedObject struct {
	ObjectName    string
	ObjectPointer interface{}
}

type Hub struct {
	Peers              map[*Peer]bool    // client's map
	PeerRegister       chan *Peer        // channel that notify peer active
	PeerUnregister     chan *Peer        // channel that notify peer inactive
	MessageChannel     chan *HubMessage  // message channel
	NamedObjectChannel chan *NamedObject // named object channel

	CallbackOnActive   func(*Peer)        // on peer active
	CallbackOnInactive func(*Peer)        // on peer inactive
	CallbackOnMsg      func(*HubMessage)  // on message
	CallbackOnObj      func(*NamedObject) // on named object
}

type HubMessage struct {
	Peer    *Peer
	Message []byte
}

func NewHub() *Hub {
	return &Hub{
		Peers:              make(map[*Peer]bool),
		PeerRegister:       make(chan *Peer),
		PeerUnregister:     make(chan *Peer),
		MessageChannel:     make(chan *HubMessage, 100),
		NamedObjectChannel: make(chan *NamedObject, 100),
		CallbackOnActive:   nil,
		CallbackOnInactive: nil,
		CallbackOnMsg:      nil,
		CallbackOnObj:      nil,
	}
}

func (this *Hub) Run() {
	for {
		select {
		case peer := <-this.PeerRegister:
			this.Peers[peer] = true
			if this.CallbackOnActive != nil {
				this.CallbackOnActive(peer)
			}
		case peer := <-this.PeerUnregister:
			if this.CallbackOnInactive != nil {
				this.CallbackOnInactive(peer)
			}
			if _, ok := this.Peers[peer]; ok {
				delete(this.Peers, peer)
				close(peer.SendChannel)
			}
		case hubMessage := <-this.MessageChannel:
			if this.CallbackOnMsg != nil {
				this.CallbackOnMsg(hubMessage)
			}
		case objMessage := <-this.NamedObjectChannel:
			if this.CallbackOnObj != nil {
				this.CallbackOnObj(objMessage)
			}
		}
	}
}
