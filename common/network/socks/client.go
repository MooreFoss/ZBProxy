package socks

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"

	"github.com/layou233/zbproxy/v3/common/network"
)

type Client struct {
	Dialer             network.Dialer
	Version            string
	Network            string
	Address            string
	Username, Password string
	Methods            []byte
}

// NewClientFromURL
//
// socks(5)://username:password@127.0.0.1:1080
//
// socks4(a)://userid@127.0.0.1:1080
func NewClientFromURL(dialer network.Dialer, s string) (*Client, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	c := &Client{Dialer: dialer}
	switch u.Scheme {
	case "socks", "socks5", "":
		c.Version = "5"
		c.Username = u.User.Username()
		c.Password, _ = u.User.Password()
	case "socks4a":
		c.Version = "4a"
		c.Username = u.User.Username()
	case "socks4":
		c.Version = "4"
		c.Username = u.User.Username()
	default:
		return nil, fmt.Errorf("socks: unknown SOCKS version: %v", u.Scheme)
	}
	c.Network = "tcp" // TODO: tcp4, tcp6 and unix support
	c.Address = u.Host
	return c, nil
}

func (c *Client) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := c.Dialer.DialContext(ctx, c.Network, c.Address)
	if err != nil {
		return nil, fmt.Errorf("socks: fail to dial to SOCKS server: %v", err)
	}
	if err = c.Handshake(conn, conn, network, address); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func (c *Client) Handshake(r io.Reader, w io.Writer, network, address string) error {
	switch c.GetVersion() {
	case "5":
		return c.handshake5(r, w, network, address)
	case "4a":
		return c.handshake4A(r, w, address)
	case "4":
		return c.handshake4(r, w, address)
	}
	return fmt.Errorf("socks: unknown SOCKS version: %v", c.Version)
}

func (c *Client) GetVersion() string {
	c.Version = strings.ToLower(c.Version)
	switch c.Version {
	case "5", "4a", "4":
		return c.Version
	case "", "socks", "socks5":
		return "5"
	case "socks4a":
		return "4a"
	case "socks4":
		return "4"
	}
	return "UNKNOWN"
}
