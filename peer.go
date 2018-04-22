package cascade

import (
	"log"

	"github.com/gorilla/websocket"
)

type Peer struct {
	Hub            *Hub            // peer's hub
	Conn           *websocket.Conn // websocket connection
	SendChannel    chan []byte     // send channel
	CallbackOnRead func([]byte)    // on read message form peer
	ExtraInfo      interface{}     // extra information
}

func NewPeer(hub *Hub, conn *websocket.Conn) *Peer {
	return &Peer{
		Hub:            hub,
		Conn:           conn,
		SendChannel:    make(chan []byte, 100),
		CallbackOnRead: nil,
		ExtraInfo:      nil,
	}
}

// peer read
func (this *Peer) ReadPump(maxReadSize int64) {
	defer func() {
		this.Hub.PeerUnregister <- this
		this.Conn.Close()
	}()

	this.Conn.SetReadLimit(maxReadSize)
	for {
		_, message, err := this.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if this.CallbackOnRead != nil {
			this.CallbackOnRead(message)
		}
	}
}

// peer write
func (this *Peer) WritePump() {
	defer func() {
		this.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-this.SendChannel:
			if !ok {
				// The hub closed the channel.
				this.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := this.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
