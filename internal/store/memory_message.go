package store

import (
	"context"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"sync"
	"time"
)

type idGen struct {
	value int
	lock  sync.Mutex
}

func (i *idGen) nextId() int {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.value++
	return i.value
}

type subjectStore struct {
	idg      idGen
	messages sync.Map
}

func (s *subjectStore) SaveMessage(message messageWithDeadline) error {
	newId := s.idg.nextId()
	message.Message.Id = newId
	s.messages.Store(newId, message)

	return nil
}

func (s *subjectStore) GetMessage(id int) (messageWithDeadline, bool) {
	m, ok := s.messages.Load(id)
	if !ok {
		return messageWithDeadline{}, ok
	}
	return m.(messageWithDeadline), ok
}

type inMemoryMessage struct {
	subjects     sync.Map
	timeProvider TimeProvider
}

type messageWithDeadline struct {
	*broker.Message
	deadline time.Time
}

func NewInMemoryMessage(provider TimeProvider) Message {
	return &inMemoryMessage{
		timeProvider: provider,
	}
}

func (i *inMemoryMessage) SaveMessage(ctx context.Context, subject string, message *broker.Message) error {
	ss := i.getSubjectStore(subject)
	return ss.SaveMessage(messageWithDeadline{
		Message:  message,
		deadline: i.timeProvider.GetCurrentTime().Add(message.Expiration),
	})
}

func (i *inMemoryMessage) GetMessage(ctx context.Context, subject string, id int) (*broker.Message, error) {
	ss := i.getSubjectStore(subject)
	currentTime := i.timeProvider.GetCurrentTime()

	message, ok := ss.GetMessage(id)
	if !ok {
		return nil, ErrInvalidId
	}

	if currentTime.After(message.deadline) {
		return nil, ErrExpired
	}

	return message.Message, nil
}

func (i *inMemoryMessage) getSubjectStore(subject string) *subjectStore {
	s, _ := i.subjects.LoadOrStore(subject, &subjectStore{})
	return s.(*subjectStore)
}
