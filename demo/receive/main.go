package main

import (
	"encoding/json"
	"errors"
	"log"
	"rabbit"

	"github.com/streadway/amqp"
)

type message struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func main() {
	err := rabbit.NewSubscriber("event", "amqp://guest:guest@localhost:5672/", "master", 3, 30, getMsg)
	if err != nil {
		log.Println(err)
	}
}

func getMsg(d amqp.Delivery) error {
	log.Printf("Received a message: %s", d.Body)

	var data message
	json.Unmarshal(d.Body, &data)

	if data.Status == "error" {
		return errors.New("模擬錯誤")
	}
	return nil
}
