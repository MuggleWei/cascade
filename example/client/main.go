package main

import (
	"log"
	"time"

	"github.com/MuggleWei/cascade/example"
	"github.com/gorilla/websocket"
)

func main() {
	addr := "ws://127.0.0.1:10102/ws"
	// addr := "ws://127.0.0.1:10000/timer"
	for {
		c, _, err := websocket.DefaultDialer.Dial(addr, nil)
		if err != nil {
			log.Printf("[Error] failed dial to %v: %v", addr, err.Error())
			time.Sleep(time.Second * 3)
			continue
		}

		loginReq := example.CommonMessage{Op: "login", Data: example.LoginMessage{User: "muggle", Password: "xxx"}}
		err = c.WriteJSON(loginReq)
		if err != nil {
			log.Println(err)
		}

		defer func() {
			c.Close()
		}()

		c.SetReadLimit(1024)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			}

			log.Println(string(message))
		}
	}
}
