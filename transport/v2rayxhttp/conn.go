package v2rayxhttp

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Conn struct {
	writer  io.WriteCloser
	reader  io.ReadCloser
	onClose func()
	NetAddr

	closeOnce  sync.Once
	closed     atomic.Bool
	closeErr   error
	deadlineMu sync.Mutex

	// deadlines
	deadline *time.Timer
}

func (c *Conn) Write(b []byte) (int, error) {
	if c.closed.Load() {
		return 0, net.ErrClosed
	}
	n, err := c.writer.Write(b)
	if err != nil && c.closed.Load() {
		err = net.ErrClosed
	}
	return n, err
}
func (c *Conn) Read(b []byte) (int, error) {
	if c.closed.Load() {
		return 0, net.ErrClosed
	}
	n, err := c.reader.Read(b)
	// Only classify errors after an explicit local Close as normal shutdown.
	// Errors returned by a still-open remote stream remain visible.
	if err != nil && c.closed.Load() {
		err = net.ErrClosed
	}
	return n, err
}
func (c *Conn) Close() error {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		c.deadlineMu.Lock()
		if c.deadline != nil {
			c.deadline.Stop()
			c.deadline = nil
		}
		c.deadlineMu.Unlock()
		if c.onClose != nil {
			c.onClose()
		}
		c.closeErr = errors.Join(c.writer.Close(), c.reader.Close())
	})
	return c.closeErr
}

func (c *Conn) SetReadDeadline(t time.Time) error  { return c.SetDeadline(t) }
func (c *Conn) SetWriteDeadline(t time.Time) error { return c.SetDeadline(t) }

func (c *Conn) SetDeadline(t time.Time) error {
	c.deadlineMu.Lock()
	defer c.deadlineMu.Unlock()
	if c.closed.Load() {
		return net.ErrClosed
	}
	if t.IsZero() {
		if c.deadline != nil {
			c.deadline.Stop()
			c.deadline = nil
		}
		return nil
	}
	d := time.Until(t)
	if c.deadline != nil {
		c.deadline.Reset(d)
		return nil
	}
	c.deadline = time.AfterFunc(d, func() {
		c.Close()
	})
	return nil
}
