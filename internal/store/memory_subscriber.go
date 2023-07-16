package store

import (
	"container/list"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"sync"
	"time"
)

const (
	publishTimeout = time.Second
)

type inMemorySubscriber struct {
	subscribers sync.Map
}

func NewInMemorySubscriber() Subscriber {
	return &inMemorySubscriber{}
}

func (i *inMemorySubscriber) AddSubscriber(subject string, callBack OnPublishFunc) {
	l, _ := i.subscribers.LoadOrStore(subject, list.New())
	l.(*list.List).PushBack(callBack)
}

func (i *inMemorySubscriber) Publish(subject string, message *broker.Message) {
	l, ok := i.subscribers.Load(subject)
	if !ok {
		return
	}
	subscribers := l.(*list.List)

	var wg sync.WaitGroup
	for element := subscribers.Front(); element != nil; element = element.Next() {
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
