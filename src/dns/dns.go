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
			ip, err := resolveDnsDomainToIp(addr, "ip4")
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

func resolveDnsDomainToIp(domain string, ipv string) (ip *net.IPAddr, err error) {
	cl := godns.Client{}
	req := godns.Msg{}
	req.SetQuestion("google.com.", godns.TypeA)
	req.SetEdns0(4096, true)
	answer, _, err := cl.Exchange(&req, "8.8.8.8:53")
	if err != nil {
		return ip, fmt.Errorf("failed to resolve with dns-resolver: %v", err)
	}
	fmt.Printf(answer.Extra[0].String())
	addr := net.IPAddr{
		IP:   net.ParseIP("1.1.1.1"),
		Zone: "",
	}
	return &addr, err
}
