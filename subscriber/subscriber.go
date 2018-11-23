package subscriber

import (
	"log"
	"rabbit/consts"
	"rabbit/lib"
	"rabbit/types"

	"github.com/streadway/amqp"
)

// New a subscriber
func New(name, url, exchange string, retryTimes int64, ttl int32, h types.Handler) error {
	// ----- Connect -----
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()

	// ----- Open Channel -----
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// ----- Declare Exchange -----
	err = lib.DeclareExange(ch, exchange, amqp.ExchangeTopic)
	if err != nil {
		return err
	}

	err = lib.DeclareExange(ch, exchange+"."+consts.FAILED, amqp.ExchangeTopic)
	if err != nil {
		return err
	}

	err = lib.DeclareExange(ch, exchange+"."+consts.RETRY, amqp.ExchangeTopic)
	if err != nil {
		return err
	}

	// ----- Declare Queue -----
	queue, err := lib.DeclareQueue(ch, consts.TASK+"."+name, nil)
	if err != nil {
		return err
	}

	failedQueue, err := lib.DeclareQueue(ch, consts.TASK+"."+name+"."+consts.FAILED, nil)
	if err != nil {
		return err
	}

	retryQueue, err := lib.DeclareQueue(ch, consts.TASK+"."+name+"."+consts.RETRY, amqp.Table{
		"x-dead-letter-exchange":    exchange,
		"x-dead-letter-routing-key": consts.RoutingKey,
		"x-message-ttl":             ttl * 1000,
	})
	if err != nil {
		return err
	}

	// ----- Binding Queue -----
	err = lib.BindQueue(ch, queue.Name, exchange, consts.RoutingKey)
	if err != nil {
		return err
	}

	err = lib.BindQueue(ch, queue.Name, exchange, queue.Name)
	if err != nil {
		return err
	}

	err = lib.BindQueue(ch, failedQueue.Name, exchange+"."+consts.FAILED, queue.Name)
	if err != nil {
		return err
	}

	err = lib.BindQueue(ch, retryQueue.Name, exchange+"."+consts.RETRY, queue.Name)
	if err != nil {
		return err
	}

	// ----- Consumer -----
	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack：設置 false 手動應答，處理接收後尚未處理就 crash 導致掉包的問題
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			err := h(d)
			if err == nil {
				ch.Ack(d.DeliveryTag, false)
				continue
			}

			if getRetryCounts(d) > retryTimes {
				log.Print("failed. send message to failed exchange")
				d.Headers = regenerateHeaders(d.Headers)
				ch.Publish(exchange+"."+consts.FAILED, queue.Name, false, false,
					amqp.Publishing{
						DeliveryMode: 2,
						Headers:      d.Headers,
						ContentType:  "application/json",
						Body:         d.Body,
					})
			} else {
				log.Print("exception. send message to retry exchange")
				d.Headers = regenerateHeaders(d.Headers)
				ch.Publish(exchange+"."+consts.RETRY, queue.Name, false, false,
					amqp.Publishing{
						DeliveryMode: 2,
						Headers:      d.Headers,
						ContentType:  "application/json",
						Body:         d.Body,
					})
			}
			ch.Ack(d.DeliveryTag, false)
		}
	}()
	log.Printf(" [*] Waiting for message. To exit press CTRL+C")
	<-forever
	return nil
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

func regenerateHeaders(headers amqp.Table) amqp.Table {
	if headers == nil {
		headers = amqp.Table{}
	}
	if _, ok := headers["x-orig-routing-key"]; !ok {
		headers["x-orig-routing-key"] = consts.RoutingKey
	}
	return headers
}
