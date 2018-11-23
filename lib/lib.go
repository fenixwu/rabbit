package lib

import (
	"github.com/streadway/amqp"
)

// DeclareQueue 宣告隊列
func DeclareQueue(ch *amqp.Channel, name string, args amqp.Table) (amqp.Queue, error) {
	return ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
}

// BindQueue 綁定隊列收發規則
func BindQueue(ch *amqp.Channel, queue, exchange, routingKey string) error {
	return ch.QueueBind(
		queue,      // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,      // no-wait
		nil,        // arguments
	)
}

// DeclareExange 宣告交換器
func DeclareExange(ch *amqp.Channel, name, exchangeType string) error {
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
