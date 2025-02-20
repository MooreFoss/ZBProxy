package proxyprotocol

import (
	"bytes"
	"fmt"
	"net/netip"
	"strconv"
	"time"

	"github.com/layou233/zbproxy/v3/common"
	"github.com/layou233/zbproxy/v3/common/bufio"
)

var (
	v1ProtocolTCP4    = []byte("TCP4")
	v1ProtocolTCP6    = []byte("TCP6")
	v1ProtocolUnknown = []byte("UNKNOWN")

	v1Separator = []byte(" ") // space
)

func readHeader1(conn *bufio.CachedConn) (*Header, error) {
	header := &Header{
		Version: Version1,
		Command: CommandProxy,
	}
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	transportProtocol, foundSeparator, err := conn.PeekUntil(v1Separator, common.CRLF)
	if err != nil {
		return nil, common.Cause("read v1 transport protocol type: ", err)
	}
	switch {
	case bytes.Equal(transportProtocol, v1ProtocolTCP4):
		header.TransportProtocol = TransportProtocolStream | TransportProtocolIPv4
	case bytes.Equal(transportProtocol, v1ProtocolTCP6):
		header.TransportProtocol = TransportProtocolStream | TransportProtocolIPv6
	case bytes.Equal(transportProtocol, v1ProtocolUnknown):
		header.TransportProtocol = transportProtocolUnspecified
		if bytes.Equal(foundSeparator, common.CRLF) {
			return header, nil
		}
	default:
		return nil, fmt.Errorf("unrecognized v1 protocol type: %x", transportProtocol)
	}

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	rawAddress, _, err := conn.PeekUntil(v1Separator)
	if err != nil {
		return nil, common.Cause("read v1 source address: ", err)
	}
	address, err := netip.ParseAddr(string(rawAddress))
	if err != nil {
		return nil, common.Cause("parse v1 source address: ", err)
	}

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, _, err = conn.PeekUntil(v1Separator)
	if err != nil {
		return nil, common.Cause("read v1 destination address: ", err)
	}
	// currently we do not handle destination address

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	rawSourcePort, _, err := conn.PeekUntil(v1Separator)
	if err != nil {
		return nil, common.Cause("read v1 source port: ", err)
	}
	sourcePort, err := strconv.ParseUint(string(rawSourcePort), 10, 16)
	if err != nil {
		return nil, common.Cause("parse v1 source port: ", err)
	}
	header.SourceAddress = netip.AddrPortFrom(address.Unmap(), uint16(sourcePort))

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, _, err = conn.PeekUntil(common.CRLF)
	if err != nil {
		return nil, common.Cause("read v1 destination port: ", err)
	}
	// currently we do not handle destination port

	return header, nil
}
