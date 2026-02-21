package proxy

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

const (
	sourceSignalTailscaleClientIP = "tailscale_client_ip"
	sourceSignalForwarded         = "forwarded"
	sourceSignalXForwardedFor     = "x_forwarded_for"
	sourceSignalXRealIP           = "x_real_ip"
	sourceSignalRemoteAddr        = "remote_addr"
	sourceSignalUnresolved        = "unresolved"
)

func resolveSourceIP(r *http.Request, preferRemoteIP bool) (netip.Addr, string, bool) {
	if preferRemoteIP {
		if addr, ok := parseIPValue(strings.TrimSpace(r.RemoteAddr)); ok {
			return addr, sourceSignalRemoteAddr, true
		}
		return netip.Addr{}, sourceSignalUnresolved, false
	}

	if addr, ok := parseIPValue(r.Header.Get("Tailscale-Client-IP")); ok {
		return addr, sourceSignalTailscaleClientIP, true
	}

	if addr, ok := parseForwardedHeader(r.Header.Get("Forwarded")); ok {
		return addr, sourceSignalForwarded, true
	}

	if addr, ok := parseIPValue(firstCSVValue(r.Header.Get("X-Forwarded-For"))); ok {
		return addr, sourceSignalXForwardedFor, true
	}

	if addr, ok := parseIPValue(r.Header.Get("X-Real-IP")); ok {
		return addr, sourceSignalXRealIP, true
	}

	if addr, ok := parseIPValue(strings.TrimSpace(r.RemoteAddr)); ok {
		return addr, sourceSignalRemoteAddr, true
	}

	return netip.Addr{}, sourceSignalUnresolved, false
}

func parseForwardedHeader(header string) (netip.Addr, bool) {
	if strings.TrimSpace(header) == "" {
		return netip.Addr{}, false
	}

	for _, entry := range strings.Split(header, ",") {
		for _, part := range strings.Split(entry, ";") {
			key, value, found := strings.Cut(strings.TrimSpace(part), "=")
			if !found || !strings.EqualFold(key, "for") {
				continue
			}

			candidate := strings.Trim(value, "\"")
			if strings.HasPrefix(candidate, "_") {
				continue
			}
			if addr, ok := parseIPValue(candidate); ok {
				return addr, true
			}
		}
	}

	return netip.Addr{}, false
}

func firstCSVValue(value string) string {
	for _, entry := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(entry)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func parseIPValue(value string) (netip.Addr, bool) {
	candidate := strings.TrimSpace(value)
	if candidate == "" {
		return netip.Addr{}, false
	}

	candidate = strings.Trim(candidate, "\"")
	if addr, err := netip.ParseAddr(candidate); err == nil {
		return addr.Unmap(), true
	}

	if strings.HasPrefix(candidate, "[") && strings.HasSuffix(candidate, "]") {
		if addr, err := netip.ParseAddr(strings.Trim(candidate, "[]")); err == nil {
			return addr.Unmap(), true
		}
	}

	if host, _, err := net.SplitHostPort(candidate); err == nil {
		if addr, err := netip.ParseAddr(host); err == nil {
			return addr.Unmap(), true
		}
		if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
			if addr, err := netip.ParseAddr(strings.Trim(host, "[]")); err == nil {
				return addr.Unmap(), true
			}
		}
	}

	return netip.Addr{}, false
}

func allowlistedEntry(addr netip.Addr, allowlist []netip.Prefix) (netip.Prefix, bool) {
	for _, entry := range allowlist {
		if entry.Contains(addr) {
			return entry, true
		}
	}
	return netip.Prefix{}, false
}
