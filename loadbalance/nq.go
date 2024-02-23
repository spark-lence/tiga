package loadbalance

import "reflect"

type NeverQueueBalance struct {
	ShortestExpectedDelayBalance
}

func NewNeverQueueBalance(endpoints []ShortestExpectedDelayEndpoint) LoadBalance {
	return &NeverQueueBalance{
		ShortestExpectedDelayBalance{
			endpoints: endpoints,
		},
	}
}
func (n *NeverQueueBalance) Select(args ...interface{}) (Endpoint, error) {
	n.lock.Lock()
	defer n.lock.Unlock()
	if len(n.endpoints) == 0 {
		return nil, ErrNoEndpoint
	}
	endpoint := n.zeroConn()
	if endpoint != nil {
		endpoint.AddActivateConnection()
		return endpoint, nil
	
	}
	min := n.endpoints[0]
	// Overhead = （ACTIVE+1）*256/Weight
	minOverhead := float64((min.ActivateConnections() + 1) * 256) / float64(min.Weight())
	for _, endpoint := range n.endpoints {
		overhead := float64((endpoint.ActivateConnections() + 1) * 256) / float64(endpoint.Weight())
		if overhead < minOverhead {
			min = endpoint
			minOverhead = overhead
		}
	}
	min.AddActivateConnection()
	return min, nil
}
func (n *NeverQueueBalance) zeroConn() ShortestExpectedDelayEndpoint {
	for _, endpoint := range n.endpoints {
		if endpoint.ActivateConnections() == 0 {
			return endpoint
		}
	}
	return nil
}

func (n *NeverQueueBalance) Name() string {
	return "NeverQueue"
}

func (n *NeverQueueBalance) Close()error {
	n.lock.Lock()
	defer n.lock.Unlock()
	for _, endpoint := range n.endpoints {
		if err:=endpoint.Close();err!=nil{
			return err
		
		}
	}
	return nil
}
func (n *NeverQueueBalance) AddEndpoint(endpoint interface{}){
	n.lock.Lock()
	defer n.lock.Unlock()
	n.endpoints = append(n.endpoints, endpoint.(ShortestExpectedDelayEndpoint))
}

func (n *NeverQueueBalance) RemoveEndpoint(endpoint Endpoint){
	n.lock.Lock()
	defer n.lock.Unlock()
	for i, ep := range n.endpoints {
		if reflect.DeepEqual(ep, endpoint) {
			n.endpoints = append(n.endpoints[:i], n.endpoints[i+1:]...)
			break
		}
	}
}
