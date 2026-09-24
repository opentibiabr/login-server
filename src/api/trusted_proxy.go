package api

import (
	"fmt"
	"net"
	"strings"
)

type trustedProxySet []*net.IPNet

func parseTrustedProxySet(entries []string) (trustedProxySet, error) {
	proxies := make(trustedProxySet, 0, len(entries))
	for _, rawEntry := range entries {
		entry := strings.TrimSpace(rawEntry)
		if entry == "" {
			continue
		}

		if ip := net.ParseIP(entry); ip != nil {
			bits := 128
			if ipv4 := ip.To4(); ipv4 != nil {
				ip = ipv4
				bits = 32
			}
			proxies = append(proxies, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
			continue
		}

		_, network, err := net.ParseCIDR(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy %q: %w", entry, err)
		}
		proxies = append(proxies, network)
	}
	return proxies, nil
}

func (proxies trustedProxySet) containsRemoteAddress(remoteAddress string) bool {
	host := remoteAddress
	if parsedHost, _, err := net.SplitHostPort(remoteAddress); err == nil {
		host = parsedHost
	}
	host = strings.Trim(host, "[]")
	if zoneIndex := strings.LastIndex(host, "%"); zoneIndex >= 0 {
		host = host[:zoneIndex]
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, proxy := range proxies {
		if proxy.Contains(ip) {
			return true
		}
	}
	return false
}
