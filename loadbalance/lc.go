package loadbalance

import (
	"reflect"
	"sync"
)

type LeastConnectionsEndpoint interface {
	Endpoint
	ActivateConnections() int
	InactiveConnections() int
	AddActivateConnection()
}
// LeastConnectionsBalance 最小连接数负载均衡
type LeastConnectionsBalance struct {
	lock      sync.Mutex
	endpoints []LeastConnectionsEndpoint
}
// NewLeastConnectionsBalance 创建一个最小连接数负载均衡器
func NewLeastConnectionsBalance(endpoints []LeastConnectionsEndpoint) LoadBalance {
	return &LeastConnectionsBalance{
		endpoints: endpoints,
		lock:      sync.Mutex{},
	}
}
func (l *LeastConnectionsBalance) AddEndpoint(endpoint interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.endpoints = append(l.endpoints, endpoint.(LeastConnectionsEndpoint))
}
func (l *LeastConnectionsBalance) Select(args ...interface{}) (Endpoint, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if len(l.endpoints) == 0 {
		return nil, ErrNoEndpoint
	}
	min := l.endpoints[0]
	// Overhead =（ Active * 256 + Inactive ）/Weight
	minOverhead := min.ActivateConnections()*256 + min.InactiveConnections()
	for _, endpoint := range l.endpoints {
		overhead := endpoint.ActivateConnections()*256 + endpoint.InactiveConnections()
		if overhead < minOverhead {
			min = endpoint
			minOverhead = overhead
		}
	}
	min.AddActivateConnection()
	return min, nil
}

func (l *LeastConnectionsBalance) RemoveEndpoint(endpoint Endpoint) {
	l.lock.Lock()
	defer l.lock.Unlock()
	for i, ep := range l.endpoints {
		if reflect.DeepEqual(ep, endpoint){
			l.endpoints = append(l.endpoints[:i], l.endpoints[i+1:]...)
			return
		}
	}
}

func (l *LeastConnectionsBalance) Name() string {
	return "LeastConnections"
}

func (l *LeastConnectionsBalance) Close() error {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, endpoint := range l.endpoints {
		if err:=endpoint.Close();err!=nil{
			return err
		}
	}
	return nil
}