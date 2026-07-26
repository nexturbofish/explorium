package guard

import (
	"net"
	"strings"
)

var privateCIDRs = []string{
	"127.0.0.0/8",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"::1/128",
	"fc00::/7",
}

func isPrivateTarget(args map[string]any) bool {
	url, ok := args["url"].(string)
	if !ok {
		return false
	}
	host := extractHost(url)
	if host == "" {
		return false
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return false
	}
	for _, ip := range ips {
		for _, cidr := range privateCIDRs {
			_, cidrNet, _ := net.ParseCIDR(cidr)
			if cidrNet.Contains(ip) {
				return true
			}
		}
	}
	return false
}

func extractHost(rawURL string) string {
	// Strip scheme
	if idx := strings.Index(rawURL, "://"); idx >= 0 {
		rawURL = rawURL[idx+3:]
	}
	// Strip path/port
	if idx := strings.IndexByte(rawURL, '/'); idx >= 0 {
		rawURL = rawURL[:idx]
	}
	if idx := strings.IndexByte(rawURL, ':'); idx >= 0 {
		rawURL = rawURL[:idx]
	}
	return rawURL
}
