package main

import (
	"encoding/json"
	"log"
	"rabbit"
)

type message struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func main() {
	send, err := rabbit.NewPublisher("event", "amqp://guest:guest@localhost:5672/", "master")
	if err != nil {
		log.Fatal(err)
	}
	defer send.Close()

	send.Publish("test1", successMsg("test1"))
	send.Publish("test2", errMsg("test2"))
}

func successMsg(name string) (data []byte) {
	return creatMsg(name, "success")
}

func errMsg(name string) (data []byte) {
	return creatMsg(name, "error")
}

func creatMsg(name, status string) []byte {
	msg := message{name, status}
	data, _ := json.Marshal(&msg)
	return data
}
