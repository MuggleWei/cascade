package cascade

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Peer struct {
	Name           string          // client's name
	Type           string          // client or server
	Hub            *Hub            // peer's hub
	Conn           *websocket.Conn // websocket connection
	SendChannel    chan []byte     // send channel
	CallbackOnRead func([]byte)    // on read message form peer
}

func NewClient(hub *Hub, conn *websocket.Conn) *Peer {
	return &Peer{
		Name:           "unknown",
		Type:           "client",
		Hub:            hub,
		Conn:           conn,
		SendChannel:    make(chan []byte),
		CallbackOnRead: nil,
	}
}

func NewServer(hub *Hub, conn *websocket.Conn, name string) *Peer {
	return &Peer{
		Name:           name,
		Type:           "server",
		Hub:            hub,
		Conn:           conn,
		SendChannel:    make(chan []byte),
		CallbackOnRead: nil,
	}
}

// peer read
func (this *Peer) ReadPump(maxReadSize int64) {
	defer func() {
		if this.Type == "server" {
			this.Hub.ServerUnregister <- this
		} else {
			this.Hub.ClientUnregister <- this
		}
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
func (this *Peer) WritePump(writeWait time.Duration) {
	defer func() {
		this.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-this.SendChannel:
			this.Conn.SetWriteDeadline(time.Now().Add(writeWait))
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
