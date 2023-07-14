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

func (i *idGen) nextOf(id int) (int, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if id > i.value {
		return 0, ErrInvalidId
	}
	return id + 1, nil
}

func (i *idGen) peekNextId() int {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.value + 1
}

type subjectStore struct {
	idg          idGen
	messages     sync.Map
	condChannels sync.Map
}

func (s *subjectStore) SaveMessage(message messageWithDeadline) error {
	newId := s.idg.nextId()
	message.Message.Id = newId
	s.messages.Store(newId, message)

	ch, ok := s.condChannels.LoadAndDelete(newId)
	if ok {
		condChan := ch.(chan struct{})
		close(condChan)
	}

	return nil
}

func (s *subjectStore) GetMessage(id int) (messageWithDeadline, bool) {
	m, ok := s.messages.Load(id)
	if !ok {
		return messageWithDeadline{}, ok
	}
	return m.(messageWithDeadline), ok
}

func (s *subjectStore) GetCondChannelOf(id int) chan struct{} {
	ch, _ := s.condChannels.LoadOrStore(id, make(chan struct{}))
	return ch.(chan struct{})
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

func (i *inMemoryMessage) GetNextMessage(ctx context.Context, subject string, previousId *int, beforeBlockCallBack func()) (*broker.Message, error) {
	if beforeBlockCallBack == nil {
		beforeBlockCallBack = func() {}
	}
	ss := i.getSubjectStore(subject)
	var once sync.Once
	defer once.Do(beforeBlockCallBack)

	var id int
	var err error
	if previousId == nil {
		id = ss.idg.peekNextId()
	} else {
		id, err = ss.idg.nextOf(*previousId)
		if err != nil {
			return nil, err
		}
	}

	message, ok := ss.GetMessage(id)
	if ok {
		return message.Message, nil
	}

	condChan := ss.GetCondChannelOf(id)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-condChan:
			message, _ = ss.GetMessage(id)
			return message.Message, nil
		default:
			once.Do(beforeBlockCallBack)
		}
	}
}

func (i *inMemoryMessage) getSubjectStore(subject string) *subjectStore {
	s, _ := i.subjects.LoadOrStore(subject, &subjectStore{})
	return s.(*subjectStore)
}
