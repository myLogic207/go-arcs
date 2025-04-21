package store

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"
)

type ObjectStore[t Object] map[string]t

// key, value store that stores state of id in values
type MappingStore map[string]map[string]map[string]bool

type Store[t Object] interface {
	// add an object with attributes attributes apply
	// returns a hash of the source as unique id
	Set(context.Context, t) (string, error)
	// removes a config by its registered id
	Remove(context.Context, string) (bool, error)
	// returns object based on id, nil if non found
	Get(context.Context, string) t
	// returns objects based on attribute
	GetByAttributes(context.Context, map[string]string) []t
	// returns all objects
	List(context.Context) []t
}

type store[t Object] struct {
	// stores objects by source-hash
	objects ObjectStore[t]
	// maps attributes to source-hashes
	mappings MappingStore
	mu       sync.RWMutex
}

func NewStore[t Object](
	objects ObjectStore[t],
	mappings MappingStore,
) Store[t] {
	if objects == nil {
		objects = make(ObjectStore[t])
	}

	if mappings == nil {
		mappings = make(MappingStore)
	}

	return &store[t]{
		objects:  objects,
		mappings: mappings,
	}
}

func (s *store[t]) Set(
	_ context.Context,
	object t,
) (string, error) {
	id := object.ID()
	s.mu.Lock()
	s.objects[id] = object

	for key, val := range object.Attributes() {
		if s.mappings[key] == nil {
			s.mappings[key] = make(map[string]map[string]bool)
		}
		if s.mappings[key][val] == nil {
			s.mappings[key][val] = make(map[string]bool)
		}
		s.mappings[key][val][id] = true
	}
	s.mu.Unlock()

	return id, nil
}

func (s *store[t]) Get(
	_ context.Context,
	id string,
) t {
	s.mu.RLock()
	object := s.objects[id]
	s.mu.RUnlock()
	return object
}

func (s *store[t]) Remove(
	_ context.Context,
	id string,
) (bool, error) {
	s.mu.RLock()
	if _, ok := s.objects[id]; !ok {
		return false, nil
	}
	s.mu.RUnlock()

	// remove existing objects only
	s.mu.Lock()
	config := s.objects[id]
	for key, val := range config.Attributes() {
		delete(s.mappings[key][val], id)
		if len(s.mappings[key]) == 0 {
			delete(s.mappings, key)
		}
	}
	delete(s.objects, id)
	s.mu.Unlock()

	return true, nil
}

func (s *store[t]) List(
	_ context.Context,
) []t {
	var objects []t
	s.mu.RLock()
	for _, object := range s.objects {
		objects = append(objects, object)
	}
	s.mu.RUnlock()
	return objects
}

func (s *store[t]) GetByAttributes(
	_ context.Context,
	attributes map[string]string,
) []t {
	var ids []string
	s.mu.RLock()
	for key, val := range attributes {
		for id, active := range s.mappings[key][val] {
			if !active {
				continue
			}
			ids = append(ids, id)
		}
	}

	objects := make([]t, len(ids))
	for i, id := range ids {
		objects[i] = s.objects[id]
	}
	s.mu.RUnlock()
	return objects
}

// same as upstream alloy to match hashing
// https://github.com/grafana/alloy/blob/main/internal/service/remotecfg/remotecfg.go
func Hash(in []byte) string {
	fnvHash := fnv.New32()
	fnvHash.Write(in)
	return fmt.Sprintf("%x", fnvHash.Sum(nil))
}
