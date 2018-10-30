package cascade

import (
	"bytes"
	"compress/flate"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Peer struct {
	Hub            *Hub            // peer's hub
	Conn           *websocket.Conn // websocket connection
	SendChannel    chan []byte     // send channel
	CallbackOnRead func([]byte)    // on read message form peer
	ExtraInfo      interface{}     // extra information
	Header         http.Header     // header
}

func NewPeer(hub *Hub, conn *websocket.Conn) *Peer {
	return &Peer{
		Hub:            hub,
		Conn:           conn,
		SendChannel:    make(chan []byte, 100),
		CallbackOnRead: nil,
		ExtraInfo:      nil,
		Header:         nil,
	}
}

// peer read
func (this *Peer) ReadPump(maxReadSize int64) {
	defer func() {
		this.Hub.PeerUnregister <- this
		this.Conn.Close()
	}()

	if maxReadSize > 0 {
		this.Conn.SetReadLimit(maxReadSize)
	}
	for {
		messageType, message, err := this.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if this.CallbackOnRead != nil {
			switch messageType {
			case websocket.TextMessage:
				// no need uncompressed
				this.CallbackOnRead(message)
			case websocket.BinaryMessage:
				// uncompressed
				text, err := GzipDecode(message)
				if err != nil {
					log.Printf("error: %v", err)
				} else {
					this.CallbackOnRead(text)
				}
			}
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

//
func GzipDecode(in []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(in))
	defer reader.Close()

	return ioutil.ReadAll(reader)

}
