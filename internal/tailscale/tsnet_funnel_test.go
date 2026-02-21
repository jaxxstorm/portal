package tailscale

import (
	"context"
	"crypto/tls"
	"net"
	"net/netip"
	"testing"

	"tailscale.com/ipn"
)

func TestFunnelSourceIPFromConnReturnsSourceForFunnelConn(t *testing.T) {
	left, right := net.Pipe()
	t.Cleanup(func() {
		_ = left.Close()
		_ = right.Close()
	})

	conn := &ipn.FunnelConn{
		Conn: left,
		Src:  netip.MustParseAddrPort("203.0.113.10:44444"),
	}

	addr, ok := funnelSourceIPFromConn(conn)
	if !ok {
		t.Fatalf("expected funnel source IP to resolve")
	}
	if got, want := addr.String(), "203.0.113.10"; got != want {
		t.Fatalf("unexpected source IP: got %q want %q", got, want)
	}
}

func TestFunnelSourceIPFromConnReturnsSourceForTLSWrappedFunnelConn(t *testing.T) {
	left, right := net.Pipe()
	t.Cleanup(func() {
		_ = left.Close()
		_ = right.Close()
	})

	funnelConn := &ipn.FunnelConn{
		Conn: left,
		Src:  netip.MustParseAddrPort("198.51.100.42:55555"),
	}
	tlsConn := tls.Client(funnelConn, &tls.Config{})

	addr, ok := funnelSourceIPFromConn(tlsConn)
	if !ok {
		t.Fatalf("expected funnel source IP to resolve")
	}
	if got, want := addr.String(), "198.51.100.42"; got != want {
		t.Fatalf("unexpected source IP: got %q want %q", got, want)
	}
}

func TestFunnelSourceIPFromConnReturnsFalseForNonFunnelConn(t *testing.T) {
	left, right := net.Pipe()
	t.Cleanup(func() {
		_ = left.Close()
		_ = right.Close()
	})

	if _, ok := funnelSourceIPFromConn(left); ok {
		t.Fatalf("expected non-funnel connection to return false")
	}
}

func TestFunnelClientIPFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), funnelClientIPContextKey{}, " 203.0.113.5 ")

	value, ok := funnelClientIPFromContext(ctx)
	if !ok {
		t.Fatalf("expected funnel client IP in context")
	}
	if got, want := value, "203.0.113.5"; got != want {
		t.Fatalf("unexpected context IP: got %q want %q", got, want)
	}
}
