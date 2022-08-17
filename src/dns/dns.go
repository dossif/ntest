package dns

import (
	"fmt"
	"net"
)

func isIp(ip string) bool {
	b := net.ParseIP(ip)
	if b == nil {
		return false
	}
	return true
}

func resolveDomainToIp(domain string, ipv string) (ip *net.IPAddr, err error) {
	ip, err = net.ResolveIPAddr(ipv, domain)
	if err != nil {
		return ip, fmt.Errorf("failed to resolve domain name %v to %v address: %v", domain, ipv, err)
	}
	return ip, err
}

func ResolveAddr(addr string) (ip *net.IPAddr, err error) {
	switch isIp(addr) {
	case true:
		return &net.IPAddr{IP: net.ParseIP(addr)}, err
	case false:
		ip, err := resolveDomainToIp(addr, "ip4")
		if err != nil {
			return ip, fmt.Errorf("failed to resolve domain: %v", err)
		}
		return ip, err
	}
	return ip, err
}
