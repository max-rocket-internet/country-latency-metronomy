# country-latency-metronomy

A golang package for measuring latency to an IP address with a fallback, "best effort" mechanism as follows:

1. Use ICMP echo to measure latency to the destination IP
2. If previous step fails:
   1. Get country code of the destination IP using [RIPE](https://www.ripe.net/) database (whois)
   2. Run a [traceroute](https://en.wikipedia.org/wiki/Traceroute) to the destination IP, record all hops
   3. Take latest successful traceroute hop that matches the country code of the destination IP, use this as a substitute destination IP
   4. Use ICMP echo to measure latency to substitute destination IP

It will measure:

- Mean Latency
- Median Latency
- P90 Latency
- Packet Loss

The package is thread safe and includes caching of whois queries so is ideal for running in goroutines to for measuring latency to many IP addresses in parallel.

A small example CLI tool is included, [main.go](main.go), to show how the package works:

```console
$ sudo go run main.go -ip-address 82.88.43.13 -debug
DEBU[0000] Ping test successful to '82.88.43.13': 190 avg
DEBU[0000] Starting latancy analysis of '82.88.43.13'

Successful: true
Destination: 82.88.43.13
Country code: it
Alternate destination:
Median latency: 185.0ms
```

And when the fallback mechanism is used:

```console
$ sudo go run main.go -ip-address 138.120.154.124 -debug
DEBU[0002] Ping failed to '138.120.154.124', no reply
DEBU[0002] Starting traceroute to '138.120.154.124'
DEBU[0026] Traceroute to '138.120.154.124' finished, 16 hops
DEBU[0026] Filtering traceroute results for country code 'us'
DEBU[0026] Found latest good hop '171.102.250.18'
DEBU[0026] Starting latancy analysis of '171.102.250.18'

Successful: true
Destination: 138.120.154.124
Country code: us
Alternate destination: 171.102.250.18
Median latency: 200.0ms
```

**It is required to run the CLI tool with root permission as it creates raw sockets**
