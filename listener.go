package throttled

import (
	"context"
	"golang.org/x/time/rate"
	"net"
	"sync"
)

func NewListener(l net.Listener) *Listener {
	return &Listener{
		Listener: l,
		limits: struct {
			global     int
			connection int
		}{global: 0, connection: 0},
		mutex: &sync.Mutex{},
		ctx:   context.Background(),
	}
}

// Listener net.Listener decorator
type Listener struct {
	net.Listener
	limits struct {
		global     int
		connection int
	}
	limiter *rate.Limiter
	mutex   *sync.Mutex
	ctx     context.Context
}

func (l *Listener) SetLimits(limitGlobal, limitPerConn int) {
	if limitGlobal > 0 {
		l.SetLimit(limitGlobal)
	}

	if limitPerConn > 0 {
		l.limits.connection = limitPerConn
	}
}

// SetLimit  sets server global limit
func (l *Listener) SetLimit(bytesPerSec int) {
	if bytesPerSec > 0 {
		l.mutex.Lock()
		l.limits.global = bytesPerSec
		l.limiter = rate.NewLimiter(rate.Limit(bytesPerSec), 32768)
		l.mutex.Unlock()
	}
}

// Accept overrides net.Listener.Accept() function to return LimitedConnection
func (l *Listener) Accept() (*LimitedConnection, error) {
	c, err := l.Listener.Accept()
	connection := LimitedConnection{
		c,
		l,
		rate.NewLimiter(rate.Limit(l.limits.connection), 32768),
		l.ctx,
	}

	return &connection, err
}
