package network

import (
	"strings"
	"syscall"
)

const (
	TCP_FASTOPEN = 15 //nolint: revive,stylecheck
)

func NewDialerControlFromOptions(option *OutboundSocketOptions) ControlFunc {
	if option == nil {
		return nil
	}
	return func(network string, address string, c syscall.RawConn) (err error) {
		err_ := c.Control(func(fd uintptr) {
			handle := syscall.Handle(fd)

			if strings.HasPrefix(network, "tcp") {
				if option.TCPFastOpen {
					err = syscall.SetsockoptInt(handle, syscall.IPPROTO_TCP, TCP_FASTOPEN, 1)
					if err != nil {
						return
					}
				}
			}
		})
		if err_ != nil {
			return err_
		}
		return err
	}
}

func NewListenerControlFromOptions(option *InboundSocketOptions) ControlFunc {
	if option == nil {
		return nil
	}
	return func(network string, address string, c syscall.RawConn) (err error) {
		err_ := c.Control(func(fd uintptr) {
			handle := syscall.Handle(fd)

			if strings.HasPrefix(network, "tcp") {
				if option.TCPFastOpen {
					err = syscall.SetsockoptInt(handle, syscall.IPPROTO_TCP, TCP_FASTOPEN, 1)
					if err != nil {
						return
					}
				}
			}
		})
		if err_ != nil {
			return err_
		}
		return err
	}
}
