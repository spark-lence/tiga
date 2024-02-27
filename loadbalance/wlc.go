package loadbalance

import (
	"reflect"
	"sync"
)

type WeightLeastConnectionsEndpoint interface {
	Endpoint
	LeastConnectionsEndpoint
	Weight() int
}
type WeightLeastConnectionsBalance struct {
	lock      sync.Mutex
	endpoints []Endpoint
}

func NewWeightLeastConnectionsBalance(endpoints []Endpoint) *WeightLeastConnectionsBalance {
	return &WeightLeastConnectionsBalance{
		endpoints: endpoints,
		lock:      sync.Mutex{},
	}
}

func (l *WeightLeastConnectionsBalance) AddEndpoint(endpoint interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.endpoints = append(l.endpoints, endpoint.(Endpoint))
}
func (l *WeightLeastConnectionsBalance) Select(_ ...interface{}) (Endpoint, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if len(l.endpoints) == 0 {
		return nil, ErrNoEndpoint
	}
	min := l.endpoints[0].(WeightLeastConnectionsEndpoint)
	// Overhead =（ Active * 256 + Inactive ）/Weight
	minRatio := float64(min.ActivateConnections()*256+min.InactiveConnections()) / float64(min.Weight())
	for _, ep := range l.endpoints {
		endpoint := ep.(WeightLeastConnectionsEndpoint)
		ratio := float64(endpoint.Weight()) / float64(endpoint.ActivateConnections()+1)

		if ratio < minRatio {
			min = endpoint
			minRatio = ratio
		}
	}
	return min, nil
}

func (r *WeightLeastConnectionsBalance) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	for _, endpoint := range r.endpoints {
		if err := endpoint.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (r *WeightLeastConnectionsBalance) GetEndpoints() []Endpoint {
	return r.endpoints
}
func (r *WeightLeastConnectionsBalance) Name() string {
	return string(WLCBalanceType)
}

func (r *WeightLeastConnectionsBalance) RemoveEndpoint(endpoint Endpoint) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for i, ep := range r.endpoints {
		if reflect.DeepEqual(ep, endpoint) {
			r.endpoints = append(r.endpoints[:i], r.endpoints[i+1:]...)
			return
		}
	}
}
