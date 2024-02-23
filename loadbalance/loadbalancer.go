package loadbalance

import (
	"context"
	"errors"
)

type Endpoint interface {
	Get(context.Context) (interface{}, error)
	Addr() string
	Close()error
}
type LoadBalance interface {
	Select(args ...interface{}) (Endpoint, error)
	AddEndpoint(endpoint interface{})
	RemoveEndpoint(endpoint Endpoint)
	Name() string
	Close() error
}

var (
	ErrNoEndpoint = errors.New("no endpoint available")
	ErrNoSourceIP = errors.New("no source ip")
)
