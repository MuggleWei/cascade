package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/MuggleWei/cascade"
)

type timestamp struct {
	Timestamp int64 `json:"ts"`
}

func runTicker(hub *cascade.Hub) {
	go func() {
		ticker := time.NewTicker(time.Second * 1)
		for t := range ticker.C {
			ms := t.UnixNano() / 1000000
			timestamp := timestamp{Timestamp: ms}
			bytes, err := json.Marshal(timestamp)
			if err != nil {
				log.Printf("[Warning] failed marshal timestamp: %+v\n", timestamp)
				continue
			}
			hub.MessageChannel <- &cascade.HubMessage{Peer: nil, Message: bytes}
		}
	}()
}
