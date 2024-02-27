package loadbalance

import (
	"reflect"
	"sync"
)

type WeightEndpoint interface {
	Endpoint
	Weight() int
	CurrentWeight() int
	SetWeight(weight int)
}

// WeightedRoundRobinBalance 加权轮询负载均衡
// 加权轮询算法是一种分配算法，它根据服务器的不同处理能力分配不同的权重。
// 基于Nginx的加权轮询算法，Nginx的加权轮询算法是根据权重来分配请求的，权重越高的服务器，每次被选中的概率越大。
// 当前节点集初始值均为零：{0,0,0}
// 所有节点的当前权重值加上设定的权重值
// 在当前节点集中选取最大权重值的节点作为命中节点
// 命中节点的当前权重值减去总权重值作为其新权重值，其他节点保持不变
type WeightedRoundRobinBalance struct {
	curIndex  int
	endpoints []Endpoint
	lock      sync.Mutex // 用于确保并发安全
}

func NewWeightedRoundRobinBalance(endpoints []Endpoint) LoadBalance {
	return &WeightedRoundRobinBalance{
		endpoints: endpoints,
		curIndex:  0,
		lock:      sync.Mutex{},
	}
}
func (r *WeightedRoundRobinBalance) GetEndpoints() []Endpoint {
	return r.endpoints
}
func (r *WeightedRoundRobinBalance) AddEndpoint(endpoint interface{}) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.endpoints = append(r.endpoints, endpoint.(WeightEndpoint))
}

func (r *WeightedRoundRobinBalance) Select(args ...interface{}) (Endpoint, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	max, err := r.max()
	if err != nil {
		return nil, err
	}
	sum := r.sumWeight()
	max.SetWeight(max.CurrentWeight() - sum)
	r.addWeight()
	return max, nil
}

// 获取最大权重的节点
func (r *WeightedRoundRobinBalance) max() (WeightEndpoint, error) {

	if len(r.endpoints) == 0 {
		return nil, ErrNoEndpoint
	}
	max := r.endpoints[0].(WeightEndpoint)
	for _, endpoint := range r.endpoints {
		endpoint := endpoint.(WeightEndpoint)
		if endpoint.CurrentWeight() > max.CurrentWeight() {
			max = endpoint
		}
	}
	return max, nil
}

func (r *WeightedRoundRobinBalance) sumWeight() int {
	sum := 0
	for _, endpoint := range r.endpoints {
		endpoint := endpoint.(WeightEndpoint)
		sum += endpoint.CurrentWeight()
	}
	return sum
}
func (r *WeightedRoundRobinBalance) addWeight() {
	for _, endpoint := range r.endpoints {
		endpoint := endpoint.(WeightEndpoint)
		endpoint.SetWeight(endpoint.CurrentWeight() + endpoint.Weight())
	}
}

func (r *WeightedRoundRobinBalance) RemoveEndpoint(endpoint Endpoint) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for i, ep := range r.endpoints {
		if reflect.DeepEqual(ep, endpoint){
			r.endpoints = append(r.endpoints[:i], r.endpoints[i+1:]...)
			return
		}
	}
}

func (r *WeightedRoundRobinBalance) Name() string {
	return "WeightedRoundRobin"
}

func (r *WeightedRoundRobinBalance) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	for _, endpoint := range r.endpoints {
		if err:=endpoint.Close();err!=nil{
			return err
		}
	}
	return nil
}