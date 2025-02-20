package proxyprotocol

import (
	"bytes"
	"net/netip"
	"testing"

	"github.com/layou233/zbproxy/v3/common/bufio"
)

func TestServerV2(t *testing.T) {
	tests := []struct {
		name     string
		header   []byte
		source   netip.AddrPort
		protocol uint8
		isLocal  bool
	}{
		{
			name: "TCP4 127.0.0.1",
			//                                                                                     VER  IP/TCP LENGTH
			header: []byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A, 0x21, 0x11, 0x00, 0x0C,
				// IPV4 -------------|  IPV4 ----------------|   SRC PORT   DEST PORT
				0x7F, 0x00, 0x00, 0x01, 0x7F, 0x00, 0x00, 0x01, 0xCA, 0x2B, 0x04, 0x01},
			source:   netip.MustParseAddrPort("127.0.0.1:51755"),
			protocol: TransportProtocolStream | TransportProtocolIPv4,
		},
		{
			name: "UDP4 127.0.0.1",
			//                                                                                          IP/UDP
			header: []byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A, 0x21, 0x12, 0x00, 0x0C,
				0x7F, 0x00, 0x00, 0x01, 0x7F, 0x00, 0x00, 0x01, 0xCA, 0x2B, 0x04, 0x01},
			source:   netip.MustParseAddrPort("127.0.0.1:51755"),
			protocol: TransportProtocolDatagram | TransportProtocolIPv4,
		},
		{
			name: "TCP6 Proxy for TCP4 127.0.0.1",
			//                                                                                     VER  IP/TCP   LENGTH
			header: []byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A, 0x21, 0x21, 0x00, 0x24,
				// IPV6 -------------------------------------------------------------------------------------|
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x7F, 0x00, 0x00, 0x01,
				// IPV6 -------------------------------------------------------------------------------------|   SRC PORT   DEST PORT
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x7F, 0x00, 0x00, 0x01, 0xCC, 0x4C, 0x04, 0x01},
			source:   netip.MustParseAddrPort("127.0.0.1:52300"),
			protocol: TransportProtocolStream | TransportProtocolIPv6,
		},
		{
			name: "TCP6 Maximal",
			//                                                                                     VER  IP/TCP   LENGTH
			header: []byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A, 0x21, 0x21, 0x00, 0x24,
				// IPV6 -------------------------------------------------------------------------------------|
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				// IPV6 -------------------------------------------------------------------------------------|   SRC PORT   DEST PORT
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			source:   netip.MustParseAddrPort("[FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF]:65535"),
			protocol: TransportProtocolStream | TransportProtocolIPv6,
		},
		{
			name: "TCP6 Proxy for TCP6 ::1",
			//                                                                                     VER  IP/TCP   LENGTH
			header: []byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A, 0x21, 0x21, 0x00, 0x2B,
				// IPV6 -------------------------------------------------------------------------------------|
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
				// IPV6 -------------------------------------------------------------------------------------|   SRC PORT   DEST PORT
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0xCF, 0x8F, 0x04, 0x01,
				//TLVs
				0x03, 0x00, 0x04, 0xFD, 0x16, 0xEE, 0x60},
			source:   netip.MustParseAddrPort("[::1]:53135"),
			protocol: TransportProtocolStream | TransportProtocolIPv6,
		},
		{
			name: "UDP6 Proxy for UDP6 ::1",
			//                                                                                     VER  IP/UDP   LENGTH
			header: []byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A, 0x21, 0x22, 0x00, 0x2B,
				// IPV6 -------------------------------------------------------------------------------------|
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
				// IPV6 -------------------------------------------------------------------------------------|   SRC PORT   DEST PORT
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0xCF, 0x8F, 0x04, 0x01,
				//TLVs
				0x03, 0x00, 0x04, 0xFD, 0x16, 0xEE, 0x60},
			source:   netip.MustParseAddrPort("[::1]:53135"),
			protocol: TransportProtocolDatagram | TransportProtocolIPv6,
		},
		{
			name:    "Local with no trailing bytes",
			header:  []byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A, 0x20, 0x00, 0x00, 0x00},
			isLocal: true,
		},
		{
			name: "Local with trailing bytes (TLVs)",
			header: []byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A, 0x20, 0xFF, 0x00, 0x07,
				0x03, 0x00, 0x04, 0xFD, 0x16, 0xEE, 0x60},
			isLocal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.header)
			header, err := ReadHeader(bufio.NewCachedConn(&connReader{reader}))
			if err != nil {
				t.Fatalf("failed to read header: %v", err)
			}
			if header.Version != Version2 {
				t.Fatalf("version mismatch: got=%v, expect=2", header.Version)
			}
			if tt.isLocal {
				if header.Command != CommandLocal {
					t.Fatalf("unexpected command: got=%v, expect=LOCAL", header.Command)
				}
				if tt.source.IsValid() && header.SourceAddress != tt.source {
					t.Fatalf("unexpected source address: got=%v, expect=%v", header.SourceAddress, tt.source)
				}
			} else {
				if header.TransportProtocol != tt.protocol {
					t.Fatalf("unexpected transport protocol: got=%v, expect=%v", header.TransportProtocol, tt.protocol)
				}
				if header.SourceAddress != tt.source {
					t.Fatalf("unexpected source address: got=%v, expect=%v", header.SourceAddress, tt.source)
				}
			}
		})
	}
}
