package store

import "sync"

type Store[T any] struct {
	mu   sync.RWMutex
	data map[string]T
	subs map[chan Event[T]]struct{}
}

type Event[T any] struct {
	ID    string `json:"id"`
	Value T      `json:"value"`
}

func New[T any]() *Store[T] {
	return &Store[T]{data: make(map[string]T), subs: make(map[chan Event[T]]struct{})}
}

func (s *Store[T]) Put(key string, value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	for sub := range s.subs {
		sub <- Event[T]{ID: key, Value: value}
	}
}

func (s *Store[T]) Get(key string) T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

func (s *Store[T]) Subscribe() chan Event[T] {
	s.mu.Lock()
	defer s.mu.Unlock()
	sub := make(chan Event[T])
	s.subs[sub] = struct{}{}
	return sub
}

func (s *Store[T]) Unsubscribe(sub chan Event[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subs, sub)
}
