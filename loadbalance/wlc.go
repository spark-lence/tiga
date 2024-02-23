package loadbalance

import "sync"
type WeightLeastConnectionsEndpoint interface{
	Endpoint
	LeastConnectionsEndpoint
	Weight() int

}
type WeightLeastConnectionsBalance struct {
	lock      sync.Mutex
	endpoints []WeightLeastConnectionsEndpoint
}

func NewWeightLeastConnectionsBalance(endpoints []WeightLeastConnectionsEndpoint) *WeightLeastConnectionsBalance {
	return &WeightLeastConnectionsBalance{
		endpoints: endpoints,
		lock:      sync.Mutex{},
	}
}

func (l *WeightLeastConnectionsBalance) AddEndpoint(endpoint WeightLeastConnectionsEndpoint) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.endpoints = append(l.endpoints, endpoint)
}
func (l *WeightLeastConnectionsBalance) Select() (Endpoint, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if len(l.endpoints) == 0 {
		return nil, ErrNoEndpoint
	}
	min := l.endpoints[0]
	// Overhead =（ Active * 256 + Inactive ）/Weight
	minRatio := float64(min.ActivateConnections()*256 + min.InactiveConnections()) / float64(min.Weight())
	for _, endpoint := range l.endpoints {
		ratio := float64(endpoint.Weight()) / float64(endpoint.ActivateConnections()+1)

		if ratio < minRatio {
			min = endpoint
			minRatio = ratio
		}
	}
	min.AddActivateConnection()
	return min, nil
}
