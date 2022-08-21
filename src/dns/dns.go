package dns

import (
	"fmt"
	godns "github.com/miekg/dns"
	"net"
	"strconv"
)

const dnsPort = 53

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
				return ip, fmt.Errorf("failed to resolve %v with os-resoler: %v", addr, err)
			}
			return ip, err
		default:
			ip, err := resolveDnsDomainToIp(addr, "ip4", ns)
			if err != nil {
				return ip, fmt.Errorf("failed to resolve %v with dns-resolver: %v", addr, err)
			}
			return ip, err
		}
		
	}
	return ip, err
}

func resolveOsDomainToIp(domain string, ipv string) (ip *net.IPAddr, err error) {
	ip, err = net.ResolveIPAddr(ipv, domain)
	if err != nil {
		return ip, fmt.Errorf("os error: %v", err)
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
	r, _, err := cl.Exchange(&req, net.JoinHostPort(ns, strconv.Itoa(dnsPort)))
	if err != nil {
		return ip, fmt.Errorf("failed to resolve: %v", err)
	}
	if r.Rcode != godns.RcodeSuccess {
		return ip, fmt.Errorf("bad record status: %v", r.Rcode)
	}
	for _, a := range r.Answer {
		if a, ok := a.(*godns.A); ok {
			fmt.Print("--", a.A.String())
		}
	}
	ip.IP = net.ParseIP("1.2.3.4")
	return ip, err
}
