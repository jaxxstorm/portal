package tailscale

import (
	"net"
	"testing"
)

func TestFindAvailableLocalPortFromSkipsOccupiedStartPort(t *testing.T) {
	occupied, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to reserve local port: %v", err)
	}
	defer occupied.Close()

	startPort := occupied.Addr().(*net.TCPAddr).Port
	gotPort, err := FindAvailableLocalPortFrom(startPort)
	if err != nil {
		t.Fatalf("expected to find an available port, got error: %v", err)
	}
	if gotPort == startPort {
		t.Fatalf("expected a port other than occupied start port %d", startPort)
	}
	if gotPort < startPort {
		t.Fatalf("expected resolved port >= start port, got %d < %d", gotPort, startPort)
	}
}

func TestFindAvailableLocalPortFromRejectsInvalidStartPort(t *testing.T) {
	if _, err := FindAvailableLocalPortFrom(0); err == nil {
		t.Fatalf("expected error for invalid start port")
	}
}
