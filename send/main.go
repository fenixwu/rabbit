package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"test",              // name
		amqp.ExchangeFanout, // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	type message struct {
		Name string `json:"name"`
		Num  int    `json:"num"`
	}

	for i := 0; i <= 10; i++ {
		msg := message{"YO", i}
		data, err := json.Marshal(&msg)
		failOnError(err, "Marshal error")
		publish(ch, "test", []byte(data))
	}
}

func publish(ch *amqp.Channel, exchangeName string, message []byte) {
	err := ch.Publish(
		"test", // exchange
		"",     // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: 2,
			Body:         message,
		})
	failOnError(err, "Failed to publish a message")
}
