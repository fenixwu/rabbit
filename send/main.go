package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

const (
	exchangeName      = "master"
	exchangeFaileName = "master.failed"
	exchangeRetryName = "master.retry"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// ----- Connect -----
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// ----- Open Channel -----
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// ----- Declare Exchange -----
	err = declareExange(ch, exchangeName, amqp.ExchangeTopic)
	failOnError(err, "Failed to declare an exchange")

	err = declareExange(ch, exchangeFaileName, amqp.ExchangeTopic)
	failOnError(err, "Failed to declare an exchange")

	err = declareExange(ch, exchangeRetryName, amqp.ExchangeTopic)
	failOnError(err, "Failed to declare an exchange")

	type message struct {
		Name   string `json:"name"`
		Num    int    `json:"num"`
		Status string `json:"status"`
	}

	for i := 0; i <= 10; i++ {
		s := "success"
		if i == 5 {
			s = "error"
		}
		msg := message{"WEEEEEEEEEE!", i, s}
		data, err := json.Marshal(&msg)
		failOnError(err, "Marshal error")
		publish(ch, exchangeName, []byte(data))
	}
}

func publish(ch *amqp.Channel, exchangeName string, message []byte) {
	err := ch.Publish(
		exchangeName,       // exchange
		"task.event.creat", // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			DeliveryMode: 2, // persistence
			ContentType:  "application/json",
			Body:         message,
		})
	failOnError(err, "Failed to publish a message")
}

func declareExange(ch *amqp.Channel, name, exchangeType string) error {
	return ch.ExchangeDeclare(
		name,         // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
}
