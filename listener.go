package throttled

import (
	"context"
	"net"
	"sync"

	"golang.org/x/time/rate"
)

// Listener net.Listener decorator
type Listener struct {
	net.Listener
	limits struct {
		global     int
		connection int
	}
	limiter *rate.Limiter
	mu      *sync.Mutex
	ctx     context.Context
}

// NewListener create ne instance of the listener without limits
func NewListener(l net.Listener) *Listener {
	return &Listener{
		Listener: l,
		limits: struct {
			global     int
			connection int
		}{global: 0, connection: 0},
		mu:  &sync.Mutex{},
		ctx: context.Background(),
	}
}

// SetLimits shortcut for setting both global and per-connection limits
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
		l.mu.Lock()
		l.limits.global = bytesPerSec
		l.limiter = rate.NewLimiter(rate.Limit(bytesPerSec), 32768)
		l.mu.Unlock()
	}
}

// Accept overrides net.Listener.Accept() function to return LimitedConnection
func (l *Listener) Accept() (*LimitedConnection, error) {
	c, err := l.Listener.Accept()
	limiter := rate.NewLimiter(rate.Limit(l.limits.connection), 32768)
	connection := LimitedConnection{
		c,
		l,
		limiter,
		l.ctx,
	}

	return &connection, err
}
