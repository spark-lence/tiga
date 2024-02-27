package loadbalance

import (
	"context"
	"errors"

	"github.com/spark-lence/tiga/pool"
)

type BalanceType string

const (
	RRBalanceType             BalanceType = "RoundRobin"
	WRRBalanceType            BalanceType = "WeightedRoundRobin"
	ConsistentHashBalanceType BalanceType = "ConsistentHash"
	LCBalanceType             BalanceType = "LeastConnection"
	SEDBalanceType            BalanceType = "ShortestExpectedDelay"
	WLCBalanceType            BalanceType = "WeightedLeastConnection"
	NQBalanceType             BalanceType = "NeverQueue"
)
type EndpointMeta interface{
	Addr()string
	Weight()int
}
type Endpoint interface {
	Get(context.Context) (interface{}, error)
	Addr() string
	Close() error
	Stats() pool.Stats
	AfterTransform(ctx context.Context, cn pool.Connection)
}
type LoadBalance interface {
	Select(args ...interface{}) (Endpoint, error)
	AddEndpoint(endpoint interface{})
	RemoveEndpoint(endpoint Endpoint)
	Name() string
	Close() error
	GetEndpoints() []Endpoint
}

var (
	ErrNoEndpoint = errors.New("no endpoint available")
	ErrNoSourceIP = errors.New("no source ip")
)
