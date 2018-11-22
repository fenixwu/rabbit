package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
)

const (
	taskEventQueueName      = "task.event"
	taskEventQueueFaileName = "task.event.failed"
	taskEventQueueRetryName = "task.event.retry"
	exchangeName            = "master"
	exchangeFailedName      = "master.failed"
	exchangeRetryName       = "master.retry"
	routingKey              = "task.#"
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

	err = declareExange(ch, exchangeFailedName, amqp.ExchangeTopic)
	failOnError(err, "Failed to declare an exchange")

	err = declareExange(ch, exchangeRetryName, amqp.ExchangeTopic)
	failOnError(err, "Failed to declare an exchange")

	// ----- Declare Queue -----
	taskEventQueue, err := declareQueue(ch, taskEventQueueName, nil)
	failOnError(err, "Failed to declare a queue")

	taskFailedQueue, err := declareQueue(ch, taskEventQueueFaileName, nil)
	failOnError(err, "Failed to declare a queue")

	taskRetryQueue, err := declareQueue(ch, taskEventQueueRetryName, amqp.Table{
		"x-dead-letter-exchange":    exchangeName,
		"x-dead-letter-routing-key": routingKey,
		"x-message-ttl":             int32(5 * 1000), // 5s
	})
	failOnError(err, "Failed to declare a queue")

	// ----- Binding Queue -----
	err = bindQueue(ch, taskEventQueue.Name, exchangeName, routingKey)
	failOnError(err, "Failed to bind a queue")

	err = bindQueue(ch, taskEventQueue.Name, exchangeName, taskEventQueue.Name)
	failOnError(err, "Failed to bind a queue")

	err = bindQueue(ch, taskFailedQueue.Name, exchangeFailedName, taskEventQueue.Name)
	failOnError(err, "Failed to bind a queue")

	err = bindQueue(ch, taskRetryQueue.Name, exchangeRetryName, taskEventQueue.Name)
	failOnError(err, "Failed to bind a queue")

	// ----- Consumer -----
	msgs, err := ch.Consume(
		taskEventQueue.Name, // queue
		"",                  // consumer
		false,               // auto-ack：設置 false 手動應答，處理接收後尚未處理就 crash 導致掉包的問題
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // arguments
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			type message struct {
				Name   string `json:"name"`
				Num    int    `json:"num"`
				Status string `json:"status"`
			}

			var data message
			err := json.Unmarshal(d.Body, &data)
			failOnError(err, "Unmarshal error")

			if data.Status == "error" && getRetryCounts(d) < 5 {
				time.Sleep(7 * time.Second)
			} else {
				ch.Ack(d.DeliveryTag, false)
				continue
			}

			retryCount := getRetryCounts(d)
			log.Printf("num:%v, retryCount: %v", data.Num, retryCount)

			if retryCount > 3 {
				log.Print("failed. send message to failed exchange")
				if d.Headers == nil {
					d.Headers = amqp.Table{}
				}
				if _, ok := d.Headers["x-orig-routing-key"]; !ok {
					d.Headers["x-orig-routing-key"] = routingKey
				}
				err := ch.Publish(exchangeFailedName, taskEventQueueName, false, false,
					amqp.Publishing{
						DeliveryMode: 2,
						Headers:      d.Headers,
						ContentType:  "application/json",
						Body:         d.Body,
					})
				failOnError(err, "Failed to publish a message")
			} else {
				log.Print("exception. send message to retry exchange")
				if d.Headers == nil {
					d.Headers = amqp.Table{}
				}
				if _, ok := d.Headers["x-orig-routing-key"]; !ok {
					d.Headers["x-orig-routing-key"] = routingKey
				}
				err := ch.Publish(exchangeRetryName, taskEventQueueName, false, false,
					amqp.Publishing{
						DeliveryMode: 2,
						Headers:      d.Headers,
						ContentType:  "application/json",
						Body:         d.Body,
					})
				failOnError(err, "Failed to publish a message")
			}
			ch.Ack(d.DeliveryTag, false)
		}
	}()
	log.Printf(" [*] Waiting for message. To exit press CTRL+C")
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

func getRetryCounts(d amqp.Delivery) (counts int64) {
	if d.Headers["x-death"] == nil {
		return
	}

	header := d.Headers["x-death"].([]interface{})[0]
	if v, ok := header.(amqp.Table); ok {
		counts = v["count"].(int64)
	}

	return
}
