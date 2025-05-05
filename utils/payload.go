package utils

import "sync"

type JsonPayloads map[string]interface{}

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

func (s *SafeJsonPayloads) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.data {
		delete(s.data, k)
	}
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

	copyMap := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		copyMap[k] = v
	}
	return copyMap
}

func (s *SafeJsonPayloads) GetDC(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, exists := s.data[key]
	if !exists {
		return nil, false
	}

	// Perform a shallow copy for basic types or use deep copy logic for composite types
	switch v := val.(type) {
	case map[string]interface{}:
		copyMap := make(map[string]interface{}, len(v))
		for k, val := range v {
			copyMap[k] = val // note: values aren't deep-copied
		}
		return copyMap, true
	case []interface{}:
		copySlice := make([]interface{}, len(v))
		copy(copySlice, v)
		return copySlice, true
	default:
		// For basic types (int, float64, string, bool, etc.), return as is
		return v, true
	}
}
