package loadbalance

import (
	"reflect"
	"sync"
)

type ShortestExpectedDelayEndpoint interface {
	Endpoint
	ActivateConnections() int
	Weight() int
}

type ShortestExpectedDelayBalance struct {
	lock sync.Mutex
	endpoints []Endpoint

}

func NewShortestExpectedDelayBalance(endpoints []Endpoint) LoadBalance {
	return &ShortestExpectedDelayBalance{
		endpoints: endpoints,
		lock: sync.Mutex{},
	}
}
func (l *ShortestExpectedDelayBalance) AddEndpoint(endpoint interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.endpoints = append(l.endpoints, endpoint.(ShortestExpectedDelayEndpoint))
}

func (l *ShortestExpectedDelayBalance) RemoveEndpoint(endpoint Endpoint) {
	l.lock.Lock()
	defer l.lock.Unlock()
	for i, ep := range l.endpoints {
		if reflect.DeepEqual(ep, endpoint) {
			l.endpoints = append(l.endpoints[:i], l.endpoints[i+1:]...)
			return
		}
	}
}

func (l *ShortestExpectedDelayBalance) Select(args ...interface{}) (Endpoint, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if len(l.endpoints) == 0 {
		return nil, ErrNoEndpoint
	}
	min := l.endpoints[0].(ShortestExpectedDelayEndpoint)
	// Overhead = （ACTIVE+1）*256/Weight
	minOverhead := float64((min.ActivateConnections() + 1) * 256) / float64(min.Weight())
	for _, endpoint := range l.endpoints {
		endpoint:=endpoint.(ShortestExpectedDelayEndpoint)
		overhead := float64((endpoint.ActivateConnections() + 1) * 256) / float64(endpoint.Weight())
		if overhead < minOverhead {
			min = endpoint
			minOverhead = overhead
		}
	}
	return min, nil
}
func (s *ShortestExpectedDelayBalance)GetEndpoints() []Endpoint{
	return s.endpoints
}
func (l *ShortestExpectedDelayBalance) Name() string {
	return "ShortestExpectedDelay"
}

func (l *ShortestExpectedDelayBalance) Close()error {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, endpoint := range l.endpoints {
		if err:=endpoint.Close();err!=nil{
			return err
		}
	}
	return nil
}