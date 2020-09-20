package throttled

import (
	"context"
	"golang.org/x/time/rate"
	"net"
)

// LimitedConnection decorate the net.Conn with throtlong functionality
type LimitedConnection struct {
	net.Conn
	listener *Listener
	connLim  *rate.Limiter
	ctx      context.Context
}

// SetLimit  sets connection limit
func (lc *LimitedConnection) SetLimit(bytesPerSec int) {
	if bytesPerSec > 0 {
		lc.listener.limits.connection = bytesPerSec
		lc.connLim = rate.NewLimiter(rate.Limit(bytesPerSec), 32768)
	}
}

// Write adds throttling functionality to net.Conn.Write
func (lc LimitedConnection) Write(b []byte) (n int, err error) {

	if lc.connLim == nil && lc.listener.limiter == nil {
		return lc.Conn.Write(b)
	}

	n, err = lc.Conn.Write(b)
	if err != nil {
		return n, err
	}

	if err = lc.connLim.WaitN(lc.ctx, n); err != nil {
		return n, err
	}

	if err = lc.listener.limiter.WaitN(lc.ctx, n); err != nil {
		return n, err
	}

	return n, err
}
