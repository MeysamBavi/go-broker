package store

import "github.com/MeysamBavi/go-broker/pkg/broker"

type Subscriber interface {
	AddSubscriber(subject string, callBack OnPublishFunc)
	Publish(subject string, message *broker.Message)
}

type OnPublishFunc func(message *broker.Message)
