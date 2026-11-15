package v2rayxhttp

// 从 mihomo 移植时需要的本地替代品：
// 原来依赖 github.com/metacubex/{mihomo/common/{httputils,contextutils},randv2}，
// 这些只是小工具，直接内联，避免把 mihomo 整个模块树拉进 sing-box。

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"net/http/httptrace"
)

// ---- NetAddr ----

type NetAddr struct {
	remoteAddr net.Addr
	localAddr  net.Addr
}

func (addr NetAddr) RemoteAddr() net.Addr { return addr.remoteAddr }
func (addr NetAddr) LocalAddr() net.Addr  { return addr.localAddr }

// NewAddrContext 把真实连接的地址回填到 NetAddr 里
func NewAddrContext(addr *NetAddr, ctx context.Context) context.Context {
	return httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			addr.localAddr = connInfo.Conn.LocalAddr()
			addr.remoteAddr = connInfo.Conn.RemoteAddr()
		},
	})
}

// CloseTransport 关闭空闲连接
func CloseTransport(roundTripper http.RoundTripper) {
	if tr, ok := roundTripper.(interface{ CloseIdleConnections() }); ok {
		tr.CloseIdleConnections()
	}
}

// ---- contextutils.AfterFunc ----

func afterFunc(ctx context.Context, f func()) (stop func() bool) {
	return context.AfterFunc(ctx, f)
}

// ---- randv2 ----

func randFloat64() float64 { return rand.Float64() }
func randUint32() uint32 { return rand.Uint32() }
func randUint64() uint64 { return rand.Uint64() }

// NewUUIDV4 生成 RFC4122 v4 UUID
func NewUUIDV4() string {
	var b [16]byte
	_, _ = crand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func randN(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.IntN(n)
}

func randIntN(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.IntN(n)
}
