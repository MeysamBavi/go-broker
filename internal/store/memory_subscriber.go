package store

import (
	"container/list"
	"github.com/MeysamBavi/go-broker/pkg/broker"
)

type inMemorySubscriber struct {
	subscribers map[string]*list.List
}

func NewInMemorySubscriber() Subscriber {
	return &inMemorySubscriber{
		subscribers: make(map[string]*list.List),
	}
}

func (i *inMemorySubscriber) AddSubscriber(subject string, callBack OnPublishFunc) {
	l, ok := i.subscribers[subject]
	if !ok {
		l = list.New()
		i.subscribers[subject] = l
	}

	l.PushBack(callBack)
}

func (i *inMemorySubscriber) Publish(subject string, message *broker.Message) {
	l, ok := i.subscribers[subject]
	if !ok {
		return
	}

	var wg sync.WaitGroup
	for element := l.Front(); element != nil; element = element.Next() {
		callback := element.Value.(OnPublishFunc)
		wg.Add(1)
		go func() {
			callback(message)
			wg.Done()
		}()
	}

	allDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(allDone)
	}()
	select {
	case <-time.After(publishTimeout):
		return
	case <-allDone:
		return
	}
}
