package recycler

import (
	"log"
	"rabbit/consts"
	"rabbit/lib"
	"rabbit/types"

	"github.com/streadway/amqp"
)

func New(name, url, exchange string, retryTimes int64, ttl int32, h types.Handler, isTaskDelAfterAck bool) error {
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
	err = lib.DeclareExange(ch, exchange+"."+consts.FAILED, amqp.ExchangeTopic)
	if err != nil {
		return err
	}

	// ----- Declare Queue -----
	queue, err := lib.DeclareQueue(ch, consts.TASK+"."+name+"."+consts.FAILED, nil)
	if err != nil {
		return err
	}

	// ----- Binding Queue -----
	err = lib.BindQueue(ch, queue.Name, exchange+"."+consts.FAILED, queue.Name)
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
			if err == nil && isTaskDelAfterAck {
				ch.Ack(d.DeliveryTag, false)
				continue
			}
		}
	}()
	log.Printf(" [*] Waiting for failed message. To exit press CTRL+C")
	<-forever
	return nil
}
