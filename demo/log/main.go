package main

import (
	"log"
	"rabbit"

	"github.com/streadway/amqp"
)

func main() {
	err := rabbit.NewRecycler("event", "amqp://guest:guest@localhost:5672/", "master", 3, 30, getMsg, true)
	if err != nil {
		log.Println(err)
	}
}

func getMsg(d amqp.Delivery) error {
	log.Printf("Received a failed message: %s", d.Body)
	return nil
}
