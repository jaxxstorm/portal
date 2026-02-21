package httputil

import (
	"net"
	"time"

	"github.com/pires/go-proxyproto"
)

// NewHTTPListener creates a TCP listener and optionally requires PROXY headers.
func NewHTTPListener(addr string, requireProxyProtocol bool) (net.Listener, error) {
	baseListener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	if !requireProxyProtocol {
		return baseListener, nil
	}

	return &proxyproto.Listener{
		Listener: baseListener,
		Policy: func(net.Addr) (proxyproto.Policy, error) {
			return proxyproto.REQUIRE, nil
		},
		ReadHeaderTimeout: 5 * time.Second,
	}, nil
}
