package context_conn

import (
	"context"
	"net"
	"sync"
	"time"
)

func Dial(network string, address string) (*Conn, error) {
	c, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &Conn{Conn: c}, nil
}

func New(conn net.Conn) *Conn {
	return &Conn{Conn: conn}
}

type Conn struct {
	net.Conn
	rmux sync.Mutex
	wmux sync.Mutex
}

func do(ctx context.Context, buf []byte, tf func(time.Time) error, af func([]byte) (int, error)) (int, error) {
	sig, wait := make(chan struct{}), make(chan struct{})

	defer func() {
		close(sig)
		<-wait
		_ = tf(time.Time{})
	}()

	go func() {
		select {
		case <-ctx.Done():
			_ = tf(time.Now())
		case <-sig:
		}
		close(wait)
	}()

	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	if dl, shouldTO := ctx.Deadline(); shouldTO {
		_ = tf(dl)
	}

	return af(buf)
}

func (c *Conn) Read(ctx context.Context, buf []byte) (int, error) {
	c.rmux.Lock()
	defer c.rmux.Unlock()
	return do(ctx, buf, c.Conn.SetReadDeadline, c.Conn.Read)
}

func (c *Conn) Write(ctx context.Context, buf []byte) (int, error) {
	c.wmux.Lock()
	defer c.wmux.Unlock()
	return do(ctx, buf, c.Conn.SetWriteDeadline, c.Conn.Write)
}
