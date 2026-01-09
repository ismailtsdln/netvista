package utils

import (
	"fmt"
	"net"
	"strings"
)

// ParseCIDR expands a CIDR notation into a slice of IP strings.
func ParseCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// Remove network and broadcast for common IPv4
	if len(ips) > 2 && ip.To4() != nil {
		return ips[1 : len(ips)-1], nil
	}

	return ips, nil
}

// inc increments an IP address.
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// ParseIPRange expands a range like "192.168.1.1-192.168.1.10" into a slice of IPs.
func ParseIPRange(ipRange string) ([]string, error) {
	parts := strings.Split(ipRange, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid IP range format")
	}

	start := net.ParseIP(strings.TrimSpace(parts[0]))
	end := net.ParseIP(strings.TrimSpace(parts[1]))
	if start == nil || end == nil {
		return nil, fmt.Errorf("invalid IP address in range")
	}

	var ips []string
	for ip := start; !ip.Equal(nextIP(end)); ip = nextIP(ip) {
		ips = append(ips, ip.String())
	}
	return ips, nil
}

func nextIP(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)
	inc(next)
	return next
}

// GetPortPreset returns a comma-separated list of ports for a given preset name.
func GetPortPreset(name string) string {
	switch strings.ToLower(name) {
	case "top-100":
		return "80,443,8080,8443,8000,8888,9000,9090,3000,5000,port:21,22,23,25,53,110,143,445,3389,3306,5432,6379,1521,27017" // Truncated but representative
	case "top-1000":
		// Simplified for brevity, in a real tool this would be a large list
		return "80,443,8080,8443,8000,8888,9000,9090,3000,5000,21,22,23,25,53,110,143,445,3389,3306,5432,6379,1521,27017,8001,8008,8081,8118,10000,20000"
	case "full":
		return "1-65535"
	default:
		return name
	}
}

// ResolveTargets takes a list of mixed inputs (domains, IPs, CIDRs, ranges) and returns a flat list of strings.
func ResolveTargets(inputs []string) []string {
	var resolved []string
	for _, input := range inputs {
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Check CIDR
		if strings.Contains(input, "/") {
			if ips, err := ParseCIDR(input); err == nil {
				resolved = append(resolved, ips...)
				continue
			}
		}

		// Check Range
		if strings.Contains(input, "-") && !strings.Contains(input, ".") { // simple check to avoid domain with dash
			// This is weak, let's improve: if it has dots and a dash
			if strings.Count(input, ".") >= 3 {
				if ips, err := ParseIPRange(input); err == nil {
					resolved = append(resolved, ips...)
					continue
				}
			}
		}

		resolved = append(resolved, input)
	}
	return resolved
}
