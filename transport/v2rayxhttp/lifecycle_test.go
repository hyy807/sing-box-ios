package v2rayxhttp

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type closeBody struct {
	done  chan struct{}
	once  sync.Once
	calls atomic.Int32
}

func (b *closeBody) Read(p []byte) (int, error) {
	<-b.done
	return 0, errors.New("http2: response body closed")
}
func (b *closeBody) Close() error { b.calls.Add(1); b.once.Do(func() { close(b.done) }); return nil }

type sink struct{ calls atomic.Int32 }

func (w *sink) Write(p []byte) (int, error) { return len(p), nil }
func (w *sink) Close() error                { w.calls.Add(1); return nil }
func TestLocalCloseIsIdempotent(t *testing.T) {
	b := &closeBody{done: make(chan struct{})}
	w := &sink{}
	c := &Conn{reader: b, writer: w}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _ = c.Close() }()
	}
	wg.Wait()
	if b.calls.Load() != 1 || w.calls.Load() != 1 {
		t.Fatalf("multiple close: reader=%d writer=%d", b.calls.Load(), w.calls.Load())
	}
	if _, err := c.Read(make([]byte, 1)); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("local shutdown misclassified: %v", err)
	}
}
func TestRemoteReadErrorPreserved(t *testing.T) {
	b := &closeBody{done: make(chan struct{})}
	close(b.done)
	c := &Conn{reader: b, writer: &sink{}}
	_, err := c.Read(make([]byte, 1))
	if err == nil || errors.Is(err, net.ErrClosed) {
		t.Fatalf("remote error hidden: %v", err)
	}
}

type leasedTransport struct{ closed bool }

func (l *leasedTransport) RoundTrip(*http.Request) (*http.Response, error) { return nil, io.EOF }
func (l *leasedTransport) Close() error                                    { l.closed = true; return nil }
func TestTransportLeaseReleased(t *testing.T) {
	l := &leasedTransport{}
	CloseTransport(l)
	if !l.closed {
		t.Fatal("transport lease leaked")
	}
}

// Exercise two simultaneous HTTP/2 streams, not just in-memory Close mocks.
func TestHTTP2CloseDoesNotCancelPeer(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		var b [64]byte
		for {
			n, err := r.Body.Read(b[:])
			if n > 0 {
				_, _ = w.Write(b[:n])
				w.(http.Flusher).Flush()
			}
			if err != nil {
				return
			}
		}
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()
	tr := server.Client().Transport
	client, err := NewXHTTPClient(&Config{Host: server.Listener.Addr().String(), Path: "/", Mode: "stream-one"}, func() http.RoundTripper { return tr }, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	a, err := client.Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	b, err := client.Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	echo := func(c net.Conn, s string) {
		t.Helper()
		if _, e := c.Write([]byte(s)); e != nil {
			t.Fatal(e)
		}
		p := make([]byte, len(s))
		if _, e := io.ReadFull(c, p); e != nil {
			t.Fatal(e)
		}
		if string(p) != s {
			t.Fatalf("bad echo %q", p)
		}
	}
	echo(a, "first")
	echo(b, "second")
	_ = a.Close()
	echo(b, "peer-still-alive")
}

func TestResetRevivesCanceledTransport(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		var b [64]byte
		for {
			n, err := r.Body.Read(b[:])
			if n > 0 {
				_, _ = w.Write(b[:n])
				w.(http.Flusher).Flush()
			}
			if err != nil {
				return
			}
		}
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()
	factory := func() (*Client, error) {
		return NewXHTTPClient(&Config{Host: server.Listener.Addr().String(), Path: "/", Mode: "stream-one"}, func() http.RoundTripper { return server.Client().Transport }, nil, false)
	}
	first, err := factory()
	if err != nil {
		t.Fatal(err)
	}
	tr := &Transport{inner: first, newClient: factory}
	defer tr.Close()
	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		c, err := tr.DialContext(ctx)
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		if _, err = c.Write([]byte("ok")); err != nil {
			t.Fatal(err)
		}
		b := make([]byte, 2)
		if _, err = io.ReadFull(c, b); err != nil {
			t.Fatal(err)
		}
		_ = c.Close()
		cancel()
		old := tr.inner
		if err = tr.Reset(); err != nil {
			t.Fatal(err)
		}
		if old.ctx.Err() == nil || tr.inner.ctx.Err() != nil {
			t.Fatal("reset did not renew canceled context")
		}
	}
	_ = tr.Close()
	if err := tr.Reset(); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("reset revived permanently closed transport: %v", err)
	}
}
