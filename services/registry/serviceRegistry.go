package registry

import (
	"sync"
)

type Registry struct {
	resources map[string]interface{}
	mutex     sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		resources: make(map[string]interface{}),
	}
}

func (r *Registry) Register(key string, resource interface{}) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.resources[key] = resource
}

func (r *Registry) Get(key string) (interface{}, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	resource, exists := r.resources[key]
	return resource, exists
}
