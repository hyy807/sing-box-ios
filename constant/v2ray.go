package constant

const (
	V2RayTransportTypeHTTP        = "http"
	V2RayTransportTypeWebsocket   = "ws"
	V2RayTransportTypeQUIC        = "quic"
	V2RayTransportTypeGRPC        = "grpc"
	V2RayTransportTypeHTTPUpgrade = "httpupgrade"
	// V2RayTransportTypeXHTTP 是 Xray 的 XHTTP（splithttp）传输，sing-box 官方没有，
	// 这里从 mihomo 的 transport/xhttp 移植（FlclashYF 的 x365 节点用的就是它）。
	V2RayTransportTypeXHTTP = "xhttp"
)
