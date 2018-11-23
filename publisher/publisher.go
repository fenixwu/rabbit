package publisher

import (
	"rabbit/consts"
	"rabbit/lib"

	"github.com/streadway/amqp"
)

// Publisher 發送端
type Publisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	name     string
	exchange string
}

// New a publisher
func New(name, url, exchange string) (*Publisher, error) {
	// ----- Connect -----
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	// ----- Open Channel -----
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// ----- Declare Exchange -----
	err = lib.DeclareExange(ch, exchange, amqp.ExchangeTopic)
	if err != nil {
		return nil, err
	}

	err = lib.DeclareExange(ch, exchange+"."+consts.FAILED, amqp.ExchangeTopic)
	if err != nil {
		return nil, err
	}

	err = lib.DeclareExange(ch, exchange+"."+consts.RETRY, amqp.ExchangeTopic)
	if err != nil {
		return nil, err
	}

	return &Publisher{conn, ch, name, exchange}, nil
}

// Close 關閉
func (p *Publisher) Close() {
	p.conn.Close()
	p.channel.Close()
}

// Publish 發布訊息
func (p *Publisher) Publish(action string, message []byte) error {
	return p.channel.Publish(
		p.exchange, // exchange
		consts.TASK+"."+p.name+"."+action, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: 2, // persistence
			ContentType:  "application/json",
			Body:         message,
		})
}
