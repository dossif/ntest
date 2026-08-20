// Package dns resolves a host string (which may already be an IP) to a
// *net.IPAddr, either via the OS resolver or via a specific DNS server
// queried directly. IPv4 (A records) only.
package dns

import (
	"fmt"
	"net"
	"strconv"

	godns "github.com/miekg/dns"
)

// dnsPort is the standard port a DNS server is queried on; ns is always
// treated as a bare IP/host, never host:port.
const dnsPort = 53

// isIP reports whether s already parses as an IP literal.
func isIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// ResolveAddr resolves addr to an IPv4 address. If addr is already an IP
// literal it is returned as-is (no lookup). Otherwise it is resolved via
// the OS resolver when ns is empty, or via a direct query to ns otherwise.
func ResolveAddr(addr string, ns string) (*net.IPAddr, error) {
	if isIP(addr) {
		return &net.IPAddr{IP: net.ParseIP(addr)}, nil
	}
	if ns == "" {
		ip, err := resolveOsDomainToIp(addr, "ip4")
		if err != nil {
			return ip, fmt.Errorf("failed to resolve %v with os-resoler: %v", addr, err)
		}
		return ip, nil
	}
	ip, err := resolveDnsDomainToIp(addr, "ip4", ns)
	if err != nil {
		return ip, fmt.Errorf("failed to resolve %v with dns-resolver: %v", addr, err)
	}
	return ip, nil
}

// resolveOsDomainToIp resolves domain using the Go/OS default resolver.
func resolveOsDomainToIp(domain string, ipv string) (ip *net.IPAddr, err error) {
	ip, err = net.ResolveIPAddr(ipv, domain)
	if err != nil {
		return ip, fmt.Errorf("os error: %v", err)
	}
	return ip, err
}

// resolveDnsDomainToIp queries ns directly over UDP for domain's A (ipv ==
// "ip4") or AAAA (ipv == "ip6") record, bypassing the OS resolver/cache.
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
	// Answer may contain multiple records (e.g. CNAME + A); keep the last A
	// record seen, which is enough for our single-address use case.
	for _, a := range r.Answer {
		if a, ok := a.(*godns.A); ok {
			ip = &net.IPAddr{IP: a.A}
		}
	}
	if ip == nil {
		return ip, fmt.Errorf("no A record found for %v", domain)
	}
	return ip, nil
}
