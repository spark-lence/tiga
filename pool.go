package tiga

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrClosed performs any operation on the closed client will return this error.
	ErrClosed = errors.New("connection is closed")

	// ErrPoolExhausted is returned from a pool connection method
	// when the maximum number of database connections in the pool has been reached.
	ErrPoolExhausted = errors.New("connection pool exhausted")

	// ErrPoolTimeout timed out waiting to get a connection from the connection pool.
	ErrPoolTimeout = errors.New("connection pool timeout")

	ErrBadConn = errors.New("bad connection")
)

var timers = sync.Pool{
	New: func() interface{} {
		t := time.NewTimer(time.Hour)
		t.Stop()
		return t
	},
}

// Stats contains pool state information and accumulated stats.
type Stats struct {
	Hits     uint32 // number of times free connection was found in the pool
	Misses   uint32 // number of times free connection was NOT found in the pool
	Timeouts uint32 // number of times a wait timeout occurred

	TotalConns uint32 // number of total connections in the pool
	IdleConns  uint32 // number of idle connections in the pool
	StaleConns uint32 // number of stale connections removed from the pool
	InvalidConns uint32 // number of invalid connections removed from the pool
}
type Connection interface {
	Close() error
	Validate() bool
	ConnInstance() Conn
}
type Pool interface {
	// NewConn(context.Context) (Connection, error)
	CloseConn(Connection) error

	Get(context.Context) (Connection, error)
	Put(context.Context, Connection)
	Remove(context.Context, Connection, error)

	Len() int
	IdleLen() int
	Stats() *Stats

	Close() error
}
type ConnUsedHook func(Connection) error

type PoolOptions struct {
	Dialer             func(context.Context) (Connection, error)
	ConnectionUsedHook []ConnUsedHook
	PoolFIFO           bool
	PoolSize           int
	PoolTimeout        time.Duration
	MinIdleConns       int
	MaxIdleConns       int
	MaxActiveConns     int
}

type lastDialErrorWrap struct {
	err error
}

type ConnPool struct {
	cfg *PoolOptions

	dialErrorsNum uint32 // atomic
	lastDialError atomic.Value

	queue chan struct{}

	connsMu sync.Mutex
	// 存储所有的连接，包括空闲和正在使用的
	conns []Connection
	// 存储空闲的连接
	idleConns []Connection
	// 连接池的大小
	poolSize int
	// 空闲连接的数量
	idleConnsLen int

	stats Stats

	_closed uint32 // atomic
}

func NewPoolOptions(Dialer func(context.Context) (Connection, error)) *PoolOptions {
	return &PoolOptions{
		ConnectionUsedHook: make([]ConnUsedHook, 0),
		Dialer:             Dialer,
		PoolSize:           10,
		PoolTimeout:        6 * time.Second,
		MinIdleConns:       3,
		MaxIdleConns:       6,
		MaxActiveConns:     10,
	}

}
func NewConnPool(opt *PoolOptions) *ConnPool {
	p := &ConnPool{
		cfg: opt,

		queue:     make(chan struct{}, opt.PoolSize),
		conns:     make([]Connection, 0, opt.PoolSize),
		idleConns: make([]Connection, 0, opt.PoolSize),
	}

	p.connsMu.Lock()
	p.checkMinIdleConns()
	p.connsMu.Unlock()

	return p
}

// checkMinIdleConns 检查连接池内的连接数量是否达到最小空闲连接数
// 如果没有达到最小空闲连接数，就创建新的连接
// 如果达到了最小空闲连接数，就不创建新的连接
func (p *ConnPool) checkMinIdleConns() {
	if p.cfg.MinIdleConns == 0 {
		return
	}
	// 填充空闲连接
	for p.poolSize < p.cfg.PoolSize && p.idleConnsLen < p.cfg.MinIdleConns {
		select {
		case p.queue <- struct{}{}:
			p.poolSize++
			p.idleConnsLen++

			go func() {
				err := p.addIdleConn()
				if err != nil && err != ErrClosed {
					p.connsMu.Lock()
					p.poolSize--
					p.idleConnsLen--
					p.connsMu.Unlock()
				}

				p.freeTurn()
			}()
		default:
			return
		}
	}
}

// addIdleConn 添加一个空闲连接
func (p *ConnPool) addIdleConn() error {
	cn, err := p.dialConn(context.TODO())
	if err != nil {
		return err
	}

	p.connsMu.Lock()
	defer p.connsMu.Unlock()

	// It is not allowed to add new connections to the closed connection pool.
	if p.closed() {
		_ = cn.Close()
		return ErrClosed
	}

	p.conns = append(p.conns, cn)
	p.idleConns = append(p.idleConns, cn)
	return nil
}

// func (p *ConnPool) NewConn(ctx context.Context) (Connection, error) {
// 	return p.newConn(ctx)
// }

func (p *ConnPool) newConn(ctx context.Context) (Connection, error) {
	if p.closed() {
		return nil, ErrClosed
	}

	p.connsMu.Lock()
	if p.cfg.MaxActiveConns > 0 && p.poolSize >= p.cfg.MaxActiveConns {
		p.connsMu.Unlock()
		return nil, ErrPoolExhausted
	}
	p.connsMu.Unlock()

	cn, err := p.dialConn(ctx)
	if err != nil {
		return nil, err
	}

	p.connsMu.Lock()
	defer p.connsMu.Unlock()
	// 连接池已经满了
	if p.cfg.MaxActiveConns > 0 && p.poolSize >= p.cfg.MaxActiveConns {
		_ = cn.Close()
		return nil, ErrPoolExhausted
	}

	p.conns = append(p.conns, cn)
	p.poolSize++
	// if pooled {
	// 	// If pool is full remove the cn on next Put.
	// 	if p.poolSize >= p.cfg.PoolSize {
	// 		cn.pooled = false
	// 	} else {
	// 		p.poolSize++
	// 	}
	// }

	return cn, nil
}

func (p *ConnPool) dialConn(ctx context.Context) (Connection, error) {
	if p.closed() {
		return nil, ErrClosed
	}

	if atomic.LoadUint32(&p.dialErrorsNum) >= uint32(p.cfg.PoolSize) {
		return nil, p.getLastDialError()
	}
	// 构建新的连接
	netConn, err := p.cfg.Dialer(ctx)
	if err != nil {
		p.setLastDialError(err)
		if atomic.AddUint32(&p.dialErrorsNum, 1) == uint32(p.cfg.PoolSize) {
			go p.tryDial()
		}
		return nil, err
	}

	// cn := NewConn(netConn)
	// cn.pooled = pooled
	return netConn, nil
}

func (p *ConnPool) tryDial() {
	for {
		if p.closed() {
			return
		}

		conn, err := p.cfg.Dialer(context.Background())
		if err != nil {
			p.setLastDialError(err)
			time.Sleep(time.Second)
			continue
		}

		atomic.StoreUint32(&p.dialErrorsNum, 0)
		_ = conn.Close()
		return
	}
}

func (p *ConnPool) setLastDialError(err error) {
	p.lastDialError.Store(&lastDialErrorWrap{err: err})
}

func (p *ConnPool) getLastDialError() error {
	err, _ := p.lastDialError.Load().(*lastDialErrorWrap)
	if err != nil {
		return err.err
	}
	return nil
}

// Get 获取或创建一个连接
func (p *ConnPool) Get(ctx context.Context) (Connection, error) {
	if p.closed() {
		return nil, ErrClosed
	}
	// 排队等待获取连接的机会
	if err := p.waitTurn(ctx); err != nil {
		return nil, err
	}

	for {
		p.connsMu.Lock()
		cn, err := p.popIdle()
		p.connsMu.Unlock()

		if err != nil {
			p.freeTurn()
			return nil, err
		}

		if cn == nil {
			break
		}


		// 检查连接是否有效
		if !cn.Validate() {
			atomic.AddUint32(&p.stats.InvalidConns, 1)
			_ = p.CloseConn(cn)
			continue
		}

		atomic.AddUint32(&p.stats.Hits, 1)
		for _, hook := range p.cfg.ConnectionUsedHook {
			if err := hook(cn); err != nil {
				return nil, err
			}
		}
		return cn, nil
	}

	atomic.AddUint32(&p.stats.Misses, 1)

	newcn, err := p.newConn(ctx)
	if err != nil {
		p.freeTurn()
		return nil, err
	}

	return newcn, nil
}

// waitTurn 排队等待获取连接的机会
// 如果上下文被取消，会清理定时器并返回上下文的错误。
// 如果能成功将空结构体发送到p.queue通道中，表示成功排队，会清理定时器并返回nil。
// 如果定时器超时，表示在配置的等待时间内未能获取到连接，会增加超时统计并返回ErrPoolTimeout错误。
func (p *ConnPool) waitTurn(ctx context.Context) error {
	// 检查请求是否已经取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	select {
	case p.queue <- struct{}{}:
		return nil
	default:
	}

	timer := timers.Get().(*time.Timer)
	timer.Reset(p.cfg.PoolTimeout)

	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		timers.Put(timer)
		return ctx.Err()
	// 排队成功，可以获取到连接
	case p.queue <- struct{}{}:
		if !timer.Stop() {
			<-timer.C
		}
		timers.Put(timer)
		return nil
	case <-timer.C:
		timers.Put(timer)
		atomic.AddUint32(&p.stats.Timeouts, 1)
		return ErrPoolTimeout
	}
}

// freeTurn 释放一个排队的机会
func (p *ConnPool) freeTurn() {
	<-p.queue
}

// popIdle 从空闲连接中弹出一个连接
func (p *ConnPool) popIdle() (Connection, error) {
	if p.closed() {
		return nil, ErrClosed
	}
	n := len(p.idleConns)
	if n == 0 {
		return nil, nil
	}

	var cn Connection
	if p.cfg.PoolFIFO {
		cn = p.idleConns[0]
		copy(p.idleConns, p.idleConns[1:])
		p.idleConns = p.idleConns[:n-1]
	} else {
		idx := n - 1
		cn = p.idleConns[idx]
		p.idleConns = p.idleConns[:idx]
	}
	p.idleConnsLen--
	p.checkMinIdleConns()
	return cn, nil
}

// Release 将连接放回连接池
func (p *ConnPool) Release(ctx context.Context, cn Connection) {
	if !cn.Validate() {
		atomic.AddUint32(&p.stats.InvalidConns, 1)
		p.Remove(ctx, cn, ErrBadConn)
		return
	}



	var shouldCloseConn bool

	p.connsMu.Lock()
	// 检查连接池的数据量关系
	if p.cfg.MaxIdleConns == 0 || p.idleConnsLen < p.cfg.MaxIdleConns {
		p.idleConns = append(p.idleConns, cn)
		p.idleConnsLen++
	} else {
		// 如果空闲连接的数量已经达到了最大空闲连接数，就关闭这个连接
		log.Printf("max idle conns reached")
		p.removeConn(cn)
		shouldCloseConn = true
	}

	p.connsMu.Unlock()

	p.freeTurn()

	if shouldCloseConn {
		_ = p.closeConn(cn)
	}
}

func (p *ConnPool) Remove(_ context.Context, cn Connection, reason error) {
	p.removeConnWithLock(cn)
	p.freeTurn()
	_ = p.closeConn(cn)
}

func (p *ConnPool) CloseConn(cn Connection) error {
	p.removeConnWithLock(cn)
	return p.closeConn(cn)
}

func (p *ConnPool) removeConnWithLock(cn Connection) {
	p.connsMu.Lock()
	defer p.connsMu.Unlock()
	p.removeConn(cn)
}

// removeConn 从连接池中移除一个连接
func (p *ConnPool) removeConn(cn Connection) {
	for i, c := range p.conns {
		if c == cn {
			p.conns = append(p.conns[:i], p.conns[i+1:]...)
			// if cn.pooled {
			p.poolSize--
			p.checkMinIdleConns()
			// }
			break
		}
	}
	atomic.AddUint32(&p.stats.StaleConns, 1)
}

// CloseConn 关闭连接
func (p *ConnPool) closeConn(cn Connection) error {
	return cn.Close()
}

// Len returns total number of connections.
func (p *ConnPool) Len() int {
	p.connsMu.Lock()
	n := len(p.conns)
	p.connsMu.Unlock()
	return n
}

// IdleLen returns number of idle connections.
func (p *ConnPool) IdleLen() int {
	p.connsMu.Lock()
	n := p.idleConnsLen
	p.connsMu.Unlock()
	return n
}

func (p *ConnPool) Stats() *Stats {
	return &Stats{
		Hits:     atomic.LoadUint32(&p.stats.Hits),
		Misses:   atomic.LoadUint32(&p.stats.Misses),
		Timeouts: atomic.LoadUint32(&p.stats.Timeouts),

		TotalConns: uint32(p.Len()),
		IdleConns:  uint32(p.IdleLen()),
		StaleConns: atomic.LoadUint32(&p.stats.StaleConns),
		InvalidConns: atomic.LoadUint32(&p.stats.InvalidConns),
	}
}

func (p *ConnPool) closed() bool {
	return atomic.LoadUint32(&p._closed) == 1
}

// Close 关闭连接池
func (p *ConnPool) Close() error {
	if !atomic.CompareAndSwapUint32(&p._closed, 0, 1) {
		return ErrClosed
	}

	var firstErr error
	p.connsMu.Lock()
	for _, cn := range p.conns {
		if err := p.closeConn(cn); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	p.conns = nil
	p.poolSize = 0
	p.idleConns = nil
	p.idleConnsLen = 0
	p.connsMu.Unlock()

	return firstErr
}
