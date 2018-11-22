package main

import (
	"log"

	"github.com/streadway/amqp"
)

const (
	taskEventQueueFaileName = "task.event.failed"
	exchangeFailedName      = "master.failed"
	routingKey              = "task.*.failed"
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
	err = declareExange(ch, exchangeFailedName, amqp.ExchangeTopic)
	failOnError(err, "Failed to declare an exchange")

	// ----- Declare Queue -----
	taskFailedQueue, err := declareQueue(ch, taskEventQueueFaileName, nil)
	failOnError(err, "Failed to declare a queue")

	// ----- Binding Queue -----
	err = bindQueue(ch, taskFailedQueue.Name, exchangeFailedName, taskEventQueueFaileName)
	failOnError(err, "Failed to bind a queue")

	// ----- Consumer -----
	msgs, err := ch.Consume(
		taskFailedQueue.Name, // queue
		"",                   // consumer
		false,                // auto-ack：設置 false 手動應答，處理接收後尚未處理就 crash 導致掉包的問題
		false,                // exclusive
		false,                // no-local
		false,                // no-wait
		nil,                  // arguments
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a failed message: %s", d.Body)
			// ch.Ack(d.DeliveryTag, false)
		}
	}()
	log.Printf(" [*] Waiting for failed message. To exit press CTRL+C")
	<-forever
}

func declareQueue(ch *amqp.Channel, name string, args amqp.Table) (amqp.Queue, error) {
	return ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
}

func bindQueue(ch *amqp.Channel, queue, exchange, routingKey string) error {
	return ch.QueueBind(
		queue,      // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,      // no-wait
		nil,        // arguments
	)
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
