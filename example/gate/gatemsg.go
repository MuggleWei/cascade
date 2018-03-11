package main

type gateMessage struct {
	Command string      `json:"command"`
	Data    interface{} `json:"data"`
}
