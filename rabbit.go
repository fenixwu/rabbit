package rabbit

import (
	"rabbit/publisher"
	"rabbit/recycler"
	"rabbit/subscriber"
	"rabbit/types"
)

func NewSubscriber(name, url, exchange string, retryTimes int64, ttl int32, h types.Handler) error {
	return subscriber.New(name, url, exchange, retryTimes, ttl, h)
}

func NewPublisher(name, url, exchange string) (*publisher.Publisher, error) {
	return publisher.New(name, url, exchange)
}

func NewRecycler(name, url, exchange string, retryTimes int64, ttl int32, h types.Handler, isTaskDelAfterAck bool) error {
	return recycler.New(name, url, exchange, retryTimes, ttl, h, isTaskDelAfterAck)
}
