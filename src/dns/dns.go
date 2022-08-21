package dns

import (
	"fmt"
	godns "github.com/miekg/dns"
	"net"
)

func isIp(ip string) bool {
	b := net.ParseIP(ip)
	if b == nil {
		return false
	}
	return true
}

func ResolveAddr(addr string, ns string) (ip *net.IPAddr, err error) {
	switch isIp(addr) {
	case true:
		return &net.IPAddr{IP: net.ParseIP(addr)}, err
	case false:
		switch ns {
		case "":
			ip, err := resolveOsDomainToIp(addr, "ip4")
			if err != nil {
				return ip, fmt.Errorf("failed to resolve domain: %v", err)
			}
			return ip, err
		default:
			ip, err := resolveDnsDomainToIp(addr, "ip4", ns)
			if err != nil {
				return ip, fmt.Errorf("failed to resolve domain: %v", err)
			}
			return ip, err
		}
		
	}
	return ip, err
}

func resolveOsDomainToIp(domain string, ipv string) (ip *net.IPAddr, err error) {
	ip, err = net.ResolveIPAddr(ipv, domain)
	if err != nil {
		return ip, fmt.Errorf("failed to resolve with os-resolver: %v", err)
	}
	return ip, err
}

func resolveDnsDomainToIp(domain string, ipv string, ns string) (ip *net.IPAddr, err error) {
	cl := godns.Client{}
	req := godns.Msg{}
	var typeX uint16
	if ipv == "ip4" {
		typeX = godns.TypeA
	} else if ipv == "ip6" {
		typeX = godns.TypeAAAA
	} else {
		return ip, fmt.Errorf("unknown ip version: %v", err)
	}
	req.SetQuestion(fmt.Sprintf("%v.", domain), typeX)
	req.SetEdns0(4096, true)
	r, _, err := cl.Exchange(&req, fmt.Sprintf("%s:53", ns))
	if err != nil {
		return ip, fmt.Errorf("failed to resolve with dns-resolver: %v", err)
	}
	addr := net.IPAddr{
		IP:   nil,
		Zone: "",
	}
	for _, a := range r.Answer {
		if a, ok := a.(*godns.A); ok {
			addr.IP = net.ParseIP(a.A.String())
		}
	}
	return &addr, err
}
