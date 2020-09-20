package throttled

import (
	"context"
	"net"

	"golang.org/x/time/rate"
)

// LimitedConnection decorate the net.Conn with throttling functionality
type LimitedConnection struct {
	net.Conn
	listener *Listener
	limiter  *rate.Limiter
	ctx      context.Context
}

// SetLimit sets connection limit
func (lc *LimitedConnection) SetLimit(bytesPerSec int) {
	if bytesPerSec > 0 {
		limiter := rate.NewLimiter(rate.Limit(bytesPerSec), 32768)
		lc.listener.limits.connection = bytesPerSec
		lc.limiter = limiter
	}
}

// Write adds throttling functionality to net.Conn.Write
func (lc LimitedConnection) Write(b []byte) (n int, err error) {
	if lc.limiter == nil && lc.listener.limiter == nil {
		return lc.Conn.Write(b)
	}

	n, err = lc.Conn.Write(b)
	if err != nil {
		return n, err
	}

	if err = lc.limiter.WaitN(lc.ctx, n); err != nil {
		return n, err
	}
	if err = lc.listener.limiter.WaitN(lc.ctx, n); err != nil {
		return n, err
	}

	return n, err
}
