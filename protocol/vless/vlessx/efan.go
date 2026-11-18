package vlessx

// Port of the working FlClash X365 default wire format. Ordinary VLESS never
// enters this encoder. Experimental environment variants are intentionally absent.
import (
	"github.com/sagernet/sing-vmess"
	"github.com/sagernet/sing/common/buf"
	E "github.com/sagernet/sing/common/exceptions"
)

func writeEFANHeader(b *buf.Buffer, r Request) error {
	if r.Command != vmess.CommandTCP && r.Command != vmess.CommandUDP {
		return E.New("x365: unsupported command ", r.Command)
	}
	if _, err := b.Write([]byte{'X', '3', '6', '5', 1, r.Command}); err != nil {
		return err
	}
	if _, err := b.Write(r.UUID[:]); err != nil {
		return err
	}
	// sing-vmess writes big-endian port followed by VLESS address type/address.
	return vmess.AddressSerializer.WriteAddrPort(b, r.Destination)
}
