package utils

import "sync"

type SafeJsonPayloads struct {
	mu   sync.RWMutex
	data JsonPayloads
}

func NewSafeJsonPayloads() *SafeJsonPayloads {
	return &SafeJsonPayloads{
		data: make(JsonPayloads),
	}
}

func (s *SafeJsonPayloads) Get(key string) (interface{}, bool) {
	s.mu.RLock() // Lock for reading
	defer s.mu.RUnlock()
	val, exists := s.data[key]
	return val, exists
}

func (s *SafeJsonPayloads) Set(key string, value interface{}) {
	s.mu.Lock() // Lock for writing
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *SafeJsonPayloads) Delete(key string) {
	s.mu.Lock() // Lock for writing
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *SafeJsonPayloads) GetFloat64(key string) (float64, bool) {
	s.mu.RLock() // Lock for reading
	defer s.mu.RUnlock()
	val, exists := s.data[key]
	if !exists {
		return 0, false
	}
	f, ok := val.(float64)
	return f, ok
}

func (s *SafeJsonPayloads) GetString(key string) (string, bool) {
	s.mu.RLock() // Lock for reading
	defer s.mu.RUnlock()
	if val, ok := s.data[key]; ok {
		strVal, ok := val.(string)
		return strVal, ok
	}
	return "", false
}

func (s *SafeJsonPayloads) GetData() map[string]interface{} {
	s.mu.RLock() // Lock for reading
	defer s.mu.RUnlock()
	return s.data
}
