package v2rayxhttp

// sing-box 适配层：把 mihomo 移植过来的 xhttp 客户端接到 sing-box 的
// dialer + tls.Config（后者已包含 REALITY / uTLS / ECH）。

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sagernet/sing-box/adapter"
	sbtls "github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing-box/option"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

var _ adapter.V2RayClientTransport = (*Transport)(nil)

type Transport struct {
	inner     *Client
	mu        sync.RWMutex
	closed    bool
	newClient func() (*Client, error)
}

// NewClient 构造 XHTTP 客户端传输。
// serverAddr 是 vless 出站的目标地址，tlsConfig 由 sing-box 传入（含 REALITY）。
func NewClient(
	ctx context.Context,
	dialer N.Dialer,
	serverAddr M.Socksaddr,
	options option.V2RayXHTTPOptions,
	tlsConfig sbtls.Config,
) (adapter.V2RayClientTransport, error) {
	requestHost := options.Host
	if requestHost == "" {
		if tlsConfig != nil && tlsConfig.ServerName() != "" {
			requestHost = tlsConfig.ServerName()
		} else {
			requestHost = serverAddr.AddrString()
		}
	}
	var reuseConfig *ReuseConfig
	var keepAlivePeriod time.Duration
	if options.ReuseSettings != nil {
		reuseConfig = &ReuseConfig{
			MaxConcurrency:   options.ReuseSettings.MaxConcurrency,
			MaxConnections:   options.ReuseSettings.MaxConnections,
			CMaxReuseTimes:   options.ReuseSettings.CMaxReuseTimes,
			HMaxRequestTimes: options.ReuseSettings.HMaxRequestTimes,
			HMaxReusableSecs: options.ReuseSettings.HMaxReusableSecs,
		}
		if options.ReuseSettings.HKeepAlivePeriod > 0 {
			keepAlivePeriod = time.Duration(options.ReuseSettings.HKeepAlivePeriod) * time.Second
		}
	}

	cfg := &Config{
		Host:                 requestHost,
		Path:                 options.Path,
		Mode:                 options.Mode,
		Headers:              headerMap(options.Headers.Build()),
		NoGRPCHeader:         options.NoGRPCHeader,
		XPaddingBytes:        options.XPaddingBytes,
		XPaddingObfsMode:     options.XPaddingObfsMode,
		XPaddingKey:          options.XPaddingKey,
		XPaddingHeader:       options.XPaddingHeader,
		XPaddingPlacement:    options.XPaddingPlacement,
		XPaddingMethod:       options.XPaddingMethod,
		UplinkHTTPMethod:     options.UplinkHTTPMethod,
		SessionPlacement:     options.SessionPlacement,
		SessionKey:           options.SessionKey,
		SessionTable:         options.SessionTable,
		SessionLength:        options.SessionLength,
		SeqPlacement:         options.SeqPlacement,
		SeqKey:               options.SeqKey,
		UplinkDataPlacement:  options.UplinkDataPlacement,
		UplinkDataKey:        options.UplinkDataKey,
		UplinkChunkSize:      options.UplinkChunkSize,
		ScMaxEachPostBytes:   options.ScMaxEachPostBytes,
		ScMinPostsIntervalMs: options.ScMinPostsIntervalMs,
		ReuseConfig:          reuseConfig,
	}

	// sing-box 的 TLS（含 REALITY）在这里注入
	if tlsConfig != nil {
		fmt.Fprintf(os.Stderr, "[xhttp] tlsConfig type=%T serverName=%q nextProtos=%v\n", tlsConfig, tlsConfig.ServerName(), tlsConfig.NextProtos())
	} else {
		fmt.Fprintln(os.Stderr, "[xhttp] tlsConfig is nil")
	}
	alpn := tlsConfig.NextProtos()
	if len(alpn) == 0 {
		alpn = []string{"h2", "http/1.1"}
	}
	makeTransport := func() http.RoundTripper {
		return NewTransport(
			func(dialCtx context.Context) (net.Conn, error) {
				return dialer.DialContext(dialCtx, N.NetworkTCP, serverAddr)
			},
			func(wrapCtx context.Context, raw net.Conn, isH2 bool) (net.Conn, error) {
				if tlsConfig == nil {
					return raw, nil
				}
				// 关键：h2 传输必须把 ALPN 锁成 h2，否则协商到 http/1.1 会得到 403
				// （对齐 mihomo 的 streamTLSConn：isH2 => NextProtos = ["h2"]）
				clone := tlsConfig.Clone()
				if isH2 {
					clone.SetNextProtos([]string{"h2"})
				} else {
					clone.SetNextProtos([]string{"http/1.1"})
				}
				return clone.Client(raw)
			},
			alpn,
			keepAlivePeriod,
		)
	}

	factory := func() (*Client, error) { return NewXHTTPClient(cfg, makeTransport, nil, tlsConfig != nil) }
	client, err := factory()
	if err != nil {
		return nil, err
	}
	return &Transport{inner: client, newClient: factory}, nil
}

func (t *Transport) DialContext(ctx context.Context) (net.Conn, error) {
	t.mu.RLock()
	client, closed := t.inner, t.closed
	t.mu.RUnlock()
	if closed {
		return nil, net.ErrClosed
	}
	return client.Dial(ctx)
}

// Reset replaces the canceled generation on a network/interface change. Active
// streams of the old network are canceled; future dials use a fresh context.
func (t *Transport) Reset() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return net.ErrClosed
	}
	next, err := t.newClient()
	if err != nil {
		t.mu.Unlock()
		return err
	}
	prev := t.inner
	t.inner = next
	t.mu.Unlock()
	return prev.Close()
}
func (t *Transport) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true
	client := t.inner
	t.mu.Unlock()
	return client.Close()
}

func headerMap(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k := range h {
		out[k] = h.Get(k)
	}
	return out
}
