package main

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// validateInterval rejects non-positive intervals: time.NewTicker panics on
// a zero or negative duration, so this must be caught before Execute runs,
// not discovered as a crash.
func validateInterval(interval time.Duration) error {
	if interval <= 0 {
		return fmt.Errorf("--interval must be greater than 0 (got %s)", interval)
	}
	return nil
}

// validateTimeout rejects non-positive timeouts: a zero or negative
// deadline expires before any dial/request/ping can complete, so every
// attempt would fail immediately and forever.
func validateTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("--timeout must be greater than 0 (got %s)", timeout)
	}
	return nil
}

// validatePort rejects values outside the valid TCP port range. Without
// this, an out-of-range --port (e.g. --port 1111111111) would pass through
// unnoticed and fail every single dial attempt forever instead of being
// rejected once at startup.
func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("--port must be between 1 and 65535 (got %d)", port)
	}
	return nil
}

// httpTokenChars matches the RFC 7230 "token" character set (case-
// insensitive here; validateMethod checks case separately for a clearer
// error message).
var httpTokenChars = regexp.MustCompile(`^[!#$%&'*+\-.^_` + "`" + `|~0-9A-Za-z]+$`)

// validateMethod rejects an empty, non-token, or non-uppercase HTTP method,
// matching the documented --method contract. Without the token check, a
// method like "GE T" would pass through and fail http.NewRequest on every
// single tick instead of being rejected once at startup.
func validateMethod(method string) error {
	if method == "" {
		return fmt.Errorf("--method must not be empty")
	}
	if !httpTokenChars.MatchString(method) {
		return fmt.Errorf("--method contains characters not allowed in an HTTP method token (got %q)", method)
	}
	if method != strings.ToUpper(method) {
		return fmt.Errorf("--method must be uppercase (got %q)", method)
	}
	return nil
}

// validateDomain rejects control characters (notably CR/LF) in the Host
// header override. Go's http.Client did not reject a CRLF-laced req.Host in
// testing, so this is a cheap safety net against a malformed/injected
// request rather than reliance on the standard library catching it.
func validateDomain(domain string) error {
	for _, r := range domain {
		if unicode.IsControl(r) {
			return fmt.Errorf("--domain must not contain control characters")
		}
	}
	return nil
}

// validateWarn rejects a negative RTT warning threshold — meaningless since
// an RTT can't be negative, so it would either never trigger the intended
// comparison in a sane way or always silently do so.
func validateWarn(warn time.Duration) error {
	if warn < 0 {
		return fmt.Errorf("--warn must not be negative (got %s)", warn)
	}
	return nil
}

// validateDNS rejects a --dns value that isn't a bare IP. internal/dns
// joins it directly with the DNS port (ns:53) and queries it as a server
// address — a hostname or a "host:port" string would silently fail to
// resolve with a confusing low-level error instead of a clear one.
func validateDNS(ns string) error {
	if ns == "" {
		return nil
	}
	if net.ParseIP(ns) == nil {
		return fmt.Errorf("--dns must be a valid IP address (got %q)", ns)
	}
	return nil
}
