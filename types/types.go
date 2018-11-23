package types

import "github.com/streadway/amqp"

// Handler 接到訊息之後的動作
type Handler func(delivery amqp.Delivery) error
