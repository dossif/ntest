# ntest

Network testing CLI tool. Runs ICMP ping, TCP connect, HTTP request, or
WebSocket handshake+ping checks on a fixed interval and logs the result of
every attempt until you stop it (Ctrl+C).

## Build & install

```bash
go build -trimpath -ldflags "-X github.com/dossif/ntest/cmd.appVersion=<version>" -o ntest .
```

The `ping` subcommand needs raw socket access, so an installed binary needs
the SUID bit:

```bash
sudo cp ./ntest /usr/local/bin/
sudo chown root:staff /usr/local/bin/ntest
sudo chmod u+s /usr/local/bin/ntest
```

Current version: **1.0.0**

## Usage

```
ntest <subcommand> <host> [flags]
```

`<host>` is a positional argument, not a flag — no need to type `--host`
every time.

### Subcommands

| Subcommand | Description |
|---|---|
| `ping` | ICMP ping test |
| `tcp`  | TCP connection test |
| `http` | HTTP request test |
| `ws`   | WebSocket handshake + ping test |

### Common flags

| Flag | Default | Description |
|---|---|---|
| `<host>` | required, positional | Target host (hostname or IP); for `http` — full URI |
| `--bind` | `0.0.0.0` | Local bind address |
| `--timeout` | `3s` | Per-attempt timeout, as a Go duration (`500ms`, `3s`, `1m`, ...) |
| `--interval` | `1s` | Delay between attempts, as a Go duration |
| `--dns` | — | Custom DNS server for resolving `<host>` (e.g. `8.8.8.8`); OS resolver if unset |

Duration flags (`--timeout`, `--interval`, `--warn`) take a Go-style duration
string — a number followed by a unit, e.g. `500ms`, `3s`, `1m`, `1h30m`. A
bare number without a unit (e.g. `1000`) is rejected.

### `ping`-specific

| Flag | Default | Description |
|---|---|---|
| `--warn` | `100ms` | RTT above this threshold logs as a warning instead of info (or colors the point on the graph, with `--graph`) |
| `--graph` | `false` | Redraw a live ASCII RTT chart instead of logging a line per attempt |

`--graph` replaces the usual per-attempt log lines with a full-screen chart
that's cleared and redrawn on every tick (last 60 attempts, `--warn`-exceeding
points highlighted). A gap in the chart line marks a failed attempt, but
that alone is easy to miss at a glance — so every redraw also prints an
explicit status banner (green `✓ ok` or red `✗ ping failed: <reason>`) and a
`fails` counter in the stats line, tracking the running total of failures
since the test started (not just what's visible in the 60-attempt window).
Since it repaints the terminal, `--graph` is meant for interactive use —
piping/redirecting `ping --graph` to a file will just capture a stream of
screen-clear codes, not a normal log; use it without `--graph` for that.

### `tcp`-specific

| Flag | Default | Description |
|---|---|---|
| `--port` | `80` | Target TCP port |

### `http`-specific

| Flag | Default | Description |
|---|---|---|
| `--method` | `GET` | HTTP method (must be uppercase) |
| `--domain` | — | Overrides the `Host` request header (SNI/certificate hostname are unaffected) |
| `--body` | — | Request body |

### `ws`-specific

`<host>` is the full target URI, e.g. `ws://example.com/socket` or
`wss://example.com/socket` (scheme defaults to `ws` if omitted).

| Flag | Default | Description |
|---|---|---|
| `--domain` | — | Overrides the `Host` header sent in the handshake |

A successful attempt does a full `Upgrade: websocket` HTTP handshake, then
sends a WebSocket ping frame and waits for the pong, then closes the
connection — `--timeout` covers the handshake and the ping/pong round trip
together. This checks more than `tcp` or `http`: it confirms the WebSocket
protocol itself works end to end, not just that the port is open or that
plain HTTP responds — reverse proxies, load balancers and WAFs commonly let
ordinary HTTP through while breaking the `Upgrade` handshake (a non-101
response, e.g. a normal `200`, is reported as a handshake error).

### Examples

```bash
ntest ping 1.1.1.1 --warn 50ms
ntest ping 1.1.1.1 --graph
ntest tcp example.com --port 443 --interval 500ms
ntest http https://example.com --method GET --dns 8.8.8.8 --timeout 5s
ntest ws wss://example.com/socket --interval 5s
```

Every tick runs independently: `<host>` is resolved fresh before every
single attempt (via `--dns` if given, the OS resolver otherwise), not once
at startup. If resolution fails, that tick logs `failed to resolve host`
and moves on — it never stops the test. This means you can start `ntest`
against a domain that doesn't resolve yet (e.g. DNS record not propagated,
or added to a zone you're about to publish) and just leave it running; it
will keep logging resolution failures until the host resolves, then pick
up automatically and start reporting real results, no restart needed. It
also means a later change to the DNS record (failover, new IP) takes effect
on the very next tick.

`http` and `ws` both dial the resolved IP directly for every attempt (via
their `http.Transport`'s `DialContext`, with keep-alives disabled) rather
than letting `net/http` re-resolve the hostname itself or reuse a
connection opened against a previous tick's resolution — this is what makes
`--dns` actually affect the connection, not just the log output, and what
guarantees a DNS change takes effect on the very next tick instead of
"sticking" to whatever IP an old kept-alive connection was using.

## Project structure

```
main.go                     # entry point — calls cmd.Execute()
cmd/root.go                 # Cobra root command, Execute(), appVersion
cmd/cmd_ping.go              # ping subcommand (ICMP)
cmd/cmd_tcp.go               # tcp subcommand
cmd/cmd_http.go              # http subcommand
cmd/cmd_ws.go                # ws subcommand
internal/icmp/icmp.go        # ICMP ping test implementation (used by the "ping" subcommand)
internal/tcp/tcp.go          # TCP connection test implementation
internal/http/http.go        # HTTP request test implementation
internal/ws/ws.go            # WebSocket handshake+ping test implementation
internal/dns/dns.go          # DNS resolution (OS resolver or custom nameserver)
internal/signal/signal.go    # SIGINT/SIGTERM → context cancellation
```

## Architecture

- **Cobra** — CLI framework; each subcommand lives in its own `cmd_*.go` file, registered via `init()`.
- **`Test` interface** — a single `Execute(ctx context.Context) error` method, implemented by the `icmp` (backs the `ping` subcommand), `tcp`, `http` and `ws` packages.
- Each test runs a `time.Ticker` loop and exits cleanly when `ctx.Done()` fires.
- Each tick is independent and self-contained: DNS resolution happens fresh on every tick, not once at startup, so one tick's outcome (including a resolution failure) never affects the next.
- Graceful shutdown: `internal/signal.ContextWithSignal` cancels the context on SIGINT/SIGTERM.

## Logging

Uses `logrus` with `TextFormatter` — same `"15:04:05"` timestamp format for
all four subcommands. Each log line carries `seq` (1-based
attempt counter), `dest`, and protocol-specific fields (`port` for `tcp`,
`method` for `http`).

Log levels:
- `Info` — success
- `Warn` — ICMP RTT exceeds `--warn`, HTTP 4xx
- `Error` — DNS resolution failure (any subcommand), ICMP failure, TCP failure, HTTP 5xx or network error, WebSocket handshake or ping failure

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/digineo/go-ping` | ICMP/ping raw sockets |
| `github.com/miekg/dns` | Custom DNS resolution |
| `github.com/sirupsen/logrus` | Structured logging |
| `github.com/coder/websocket` | WebSocket client (handshake + ping/pong) |
| `github.com/guptarohit/asciigraph` | ASCII RTT chart for `ping --graph` |
