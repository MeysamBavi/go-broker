package store

import (
	"context"
	"sync"
)

type memSequence struct {
	sequences sync.Map
	locks     sync.Map
}

func NewInMemorySequence() Sequence {
	return &memSequence{}
}

func (m *memSequence) lock(subject string) {
	var mu sync.Mutex
	actual, _ := m.locks.LoadOrStore(subject, &mu)
	lock := actual.(*sync.Mutex)
	lock.Lock()
}

func (m *memSequence) unlock(subject string) {
	val, _ := m.locks.Load(subject)
	lock := val.(*sync.Mutex)
	lock.Unlock()
}

func (m *memSequence) CreateNewId(_ context.Context, subject string) (int32, error) {
	m.lock(subject)
	defer m.unlock(subject)

	val, ok := m.sequences.Load(subject)
	var id int32
	if ok {
		id = val.(int32)
	} else {
		id = 0
	}
	id++
	m.sequences.Store(subject, id)

	return id, nil
}

func (m *memSequence) Load(ctx context.Context, subject string, lastId int32) error {
	m.lock(subject)
	m.unlock(subject)

	m.sequences.Store(subject, lastId)

	return nil
}
