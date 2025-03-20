package util

import (
	"sync"
	"time"
)

var Reg *Registry

type Registry struct {
	resources  map[string]interface{}
	mutex      sync.RWMutex
	wsChannels map[string][]chan []byte
	wsMutex    sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		resources:  make(map[string]interface{}),
		wsChannels: make(map[string][]chan []byte),
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

func (r *Registry) RegisterWSChannel(topic string, ch chan []byte) {
	r.wsMutex.Lock()
	defer r.wsMutex.Unlock()

	if _, exists := r.wsChannels[topic]; !exists {
		r.wsChannels[topic] = make([]chan []byte, 0)
	}
	r.wsChannels[topic] = append(r.wsChannels[topic], ch)
}

func (r *Registry) UnregisterWSChannel(topic string, ch chan []byte) {
	r.wsMutex.Lock()
	defer r.wsMutex.Unlock()

	if channels, exists := r.wsChannels[topic]; exists {
		for i, channel := range channels {
			if channel == ch {
				r.wsChannels[topic] = append(channels[:i], channels[i+1:]...)
				break
			}
		}
	}
}

func (r *Registry) BroadcastToChannel(topic string, message []byte) {
	r.wsMutex.RLock()
	var channels []chan []byte
	if chans, exists := r.wsChannels[topic]; exists && len(chans) > 0 {
		channels = make([]chan []byte, len(chans))
		copy(channels, chans)
	}
	r.wsMutex.RUnlock()

	if len(channels) == 0 {
		return
	}

	var channelsToRemove []chan []byte

	for _, ch := range channels {
		func(ch chan []byte) {
			defer func() {
				if r := recover(); r != nil {
					channelsToRemove = append(channelsToRemove, ch)
				}
			}()

			select {
			case ch <- message:
			case <-time.After(100 * time.Millisecond):
				channelsToRemove = append(channelsToRemove, ch)
			}
		}(ch)
	}

	if len(channelsToRemove) > 0 {
		go func(toRemove []chan []byte) {
			for _, ch := range toRemove {
				r.UnregisterWSChannel(topic, ch)
			}
		}(channelsToRemove)
	}
}
