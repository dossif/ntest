// Package dns resolves a host string (which may already be an IP) to a
// *net.IPAddr, either via the OS resolver or via a specific DNS server
// queried directly. IPv4 (A records) only.
package dns

import (
	"context"
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

// ResolveAddr resolves addr to an IPv4 address, bounded by ctx. If addr is
// already an IP literal it is returned as-is (no lookup, ctx unused).
// Otherwise it is resolved via the OS resolver when ns is empty, or via a
// direct query to ns otherwise.
//
// ctx matters even for the OS-resolver path: plain net.ResolveIPAddr (the
// obvious stdlib choice) takes no context and has no timeout of its own —
// if the OS-level resolution mechanism hangs (unreachable resolver in a
// broken VPN/split-DNS setup, a stuck system resolver daemon, a firewall
// silently dropping DNS packets), it blocks forever. Since callers run this
// synchronously inside their tick loop, that would hang the whole program,
// and — because SIGINT is already being intercepted into ctx cancellation
// by then — Ctrl+C would appear to do nothing. Using the context-aware
// net.Resolver avoids that: the deadline is honored even when the
// underlying platform resolver call can't itself be interrupted.
func ResolveAddr(ctx context.Context, addr string, ns string) (*net.IPAddr, error) {
	if isIP(addr) {
		return &net.IPAddr{IP: net.ParseIP(addr)}, nil
	}
	if ns == "" {
		ip, err := resolveOsDomainToIp(ctx, addr, "ip4")
		if err != nil {
			return ip, fmt.Errorf("failed to resolve %v with os-resoler: %v", addr, err)
		}
		return ip, nil
	}
	ip, err := resolveDnsDomainToIp(ctx, addr, "ip4", ns)
	if err != nil {
		return ip, fmt.Errorf("failed to resolve %v with dns-resolver: %v", addr, err)
	}
	return ip, nil
}

// resolveOsDomainToIp resolves domain using the Go/OS default resolver,
// bounded by ctx (see the ResolveAddr doc comment for why this can't be the
// simpler, context-less net.ResolveIPAddr).
func resolveOsDomainToIp(ctx context.Context, domain string, ipv string) (*net.IPAddr, error) {
	ips, err := net.DefaultResolver.LookupIP(ctx, ipv, domain)
	if err != nil {
		return nil, fmt.Errorf("os error: %v", err)
	}
	return &net.IPAddr{IP: ips[0]}, nil
}

// resolveDnsDomainToIp queries ns directly over UDP for domain's A (ipv ==
// "ip4") or AAAA (ipv == "ip6") record, bypassing the OS resolver/cache.
// Bounded by ctx via ExchangeContext — miekg/dns's plain Exchange already
// falls back to its own internal 2s dial/read/write timeout when unset, but
// that ignores our --timeout entirely; ExchangeContext makes this path
// respect the same deadline as the OS-resolver path and the actual
// ping/dial/request that follows.
func resolveDnsDomainToIp(ctx context.Context, domain string, ipv string, ns string) (ip *net.IPAddr, err error) {
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
	r, _, err := cl.ExchangeContext(ctx, &req, net.JoinHostPort(ns, strconv.Itoa(dnsPort)))
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
