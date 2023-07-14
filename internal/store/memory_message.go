package store

import (
	"context"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"time"
)

type idGen int

func (i *idGen) nextId() int {
	*i = *i + 1
	return int(*i)
}

type inMemoryMessage struct {
	subjectIdGens map[string]*idGen
	messages      map[int]messageWithDeadline
	timeProvider  TimeProvider
}

type messageWithDeadline struct {
	*broker.Message
	deadline time.Time
}

func NewInMemoryMessage(provider TimeProvider) Message {
	return &inMemoryMessage{
		subjectIdGens: make(map[string]*idGen),
		messages:      make(map[int]messageWithDeadline),
		timeProvider:  provider,
	}
}

func (i *inMemoryMessage) SaveMessage(ctx context.Context, subject string, message *broker.Message) error {
	idg, idgOk := i.subjectIdGens[subject]
	if !idgOk {
		idg = new(idGen)
		i.subjectIdGens[subject] = idg
	}

	newId := idg.nextId()
	message.Id = newId
	i.messages[newId] = messageWithDeadline{
		Message:  message,
		deadline: i.timeProvider.GetCurrentTime().Add(message.Expiration),
	}

	return nil
}

func (i *inMemoryMessage) GetMessage(ctx context.Context, subject string, id int) (*broker.Message, error) {
	currentTime := i.timeProvider.GetCurrentTime()
	message, ok := i.messages[id]
	if !ok {
		return nil, ErrInvalidId
	}

	if currentTime.After(message.deadline) {
		return nil, ErrExpired
	}

	return message.Message, nil
}
