package store

import (
	"context"
	"sync"
)

type memSequence struct {
	sequences map[string]int32
	locks     sync.Map
}

func NewInMemorySequence() Sequence {
	return &memSequence{
		sequences: make(map[string]int32),
	}
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

	val := m.sequences[subject]
	val++
	m.sequences[subject] = val

	return val, nil
}

func (m *memSequence) Load(ctx context.Context, subject string, lastId int32) error {
	m.lock(subject)
	m.unlock(subject)

	m.sequences[subject] = lastId

	return nil
}
