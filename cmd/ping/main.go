package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"potat.dev/ping/internal/pinger"
	// "github.com/yourusername/ping-utility/internal/pinger"
)

func main() {
	// Define command-line flags
	host := flag.String("host", "8.8.8.8", "Host to ping")
	count := flag.Int("count", 4, "Number of ping attempts")
	packetSize := flag.Int("size", 64, "Size of ping packet")
	timeout := flag.Duration("timeout", 2*time.Second, "Timeout for each ping")
	interval := flag.Duration("interval", 1*time.Second, "Interval between pings")
	protocol := flag.String("protocol", "icmp", "Ping protocol (icmp or udp)")
	port := flag.Int("port", 33434, "Port to use for UDP ping")

	// Parse command-line flags
	flag.Parse()

	// Validate protocol
	var pingProtocol pinger.PingProtocol
	switch *protocol {
	case "icmp":
		pingProtocol = pinger.ProtocolICMP
	case "udp":
		pingProtocol = pinger.ProtocolUDP
	default:
		fmt.Printf("Unsupported protocol: %s. Use 'icmp' or 'udp'\n", *protocol)
		os.Exit(1)
	}

	// Create pinger with configured options
	p := pinger.NewPinger(*host,
		pinger.WithCount(*count),
		pinger.WithPacketSize(*packetSize),
		pinger.WithTimeout(*timeout),
		pinger.WithInterval(*interval),
		pinger.WithProtocol(pingProtocol),
		pinger.WithPort(*port),
	)

	// Perform ping
	results, err := p.Ping()
	if err != nil {
		fmt.Printf("Ping error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	pinger.PrintResults(*host, results)
}
