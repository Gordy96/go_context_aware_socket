package context_conn

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

func TestConn_Read(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var l net.Listener
	var err error

	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		l, err = net.Listen("tcp", ":")
		assert.NoError(t, err)

		wg.Done()

		c, err := l.Accept()
		assert.NoError(t, err)

		n, err := c.Write([]byte("foo bar baz"))
		assert.NoError(t, err)
		assert.Equal(t, 11, n)
	}()

	wg.Wait()

	conn, err := Dial("tcp", l.Addr().String())
	assert.NoError(t, err)

	sctx, cancel2 := context.WithCancel(ctx)

	go func() {
		time.Sleep(2 * time.Second)
		cancel2()
	}()

	var buf = make([]byte, 20)

	n, err := conn.Read(sctx, buf)
	assert.NoError(t, err)
	assert.Equal(t, 11, n)
	assert.ElementsMatch(t, []byte("foo bar baz"), buf[:n])

	n, err = conn.Read(sctx, buf)
	assert.ErrorIs(t, err, os.ErrDeadlineExceeded)
	assert.Zero(t, n)

	err = l.Close()
	assert.NoError(t, err)
}

func TestConn_ListenRead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var l net.Listener
	var err error

	l, err = net.Listen("tcp", ":")
	assert.NoError(t, err)

	go func() {
		conn, err := net.Dial("tcp", l.Addr().String())
		assert.NoError(t, err)

		n, err := conn.Write([]byte("foo bar baz"))
		assert.NoError(t, err)
		assert.Equal(t, 11, n)
	}()

	c, err := l.Accept()
	assert.NoError(t, err)

	conn := New(c)

	var buf = make([]byte, 20)

	n, err := conn.Read(ctx, buf)
	assert.NoError(t, err)
	assert.Equal(t, 11, n)
	assert.ElementsMatch(t, []byte("foo bar baz"), buf[:n])

	sctx, cancel2 := context.WithCancel(ctx)

	go func() {
		time.Sleep(2 * time.Second)
		cancel2()
	}()

	n, err = conn.Read(sctx, buf)
	assert.ErrorIs(t, err, os.ErrDeadlineExceeded)
	assert.Zero(t, n)

	err = l.Close()
	assert.NoError(t, err)
}
