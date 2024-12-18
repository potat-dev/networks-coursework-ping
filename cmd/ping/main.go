package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"potat.dev/ping/internal/pinger"
)

func main() {
	// Define command-line flags
	host := flag.String("host", "8.8.8.8", "Host to ping")
	count := flag.Int("count", 4, "Number of ping attempts")
	packetSize := flag.Int("size", 64, "Size of ping packet")
	timeout := flag.Duration("timeout", 2*time.Second, "Timeout for each ping")
	interval := flag.Duration("interval", 1*time.Second, "Interval between pings")
	protocol := flag.String("protocol", "icmp", "Protocol to use (icmp or udp)")
	udpPort := flag.Int("udp-port", 33434, "Target UDP port for UDP ping")

	// Parse command-line flags
	flag.Parse()

	// Create pinger with configured options
	p := pinger.NewPinger(*host,
		pinger.WithCount(*count),
		pinger.WithPacketSize(*packetSize),
		pinger.WithTimeout(*timeout),
		pinger.WithInterval(*interval),
		pinger.WithUDPPort(*udpPort),
	)

	// Perform ping
	var results []pinger.PingResult
	var err error

	switch *protocol {
	case "icmp":
		results, err = p.Ping()
	case "udp":
		results, err = p.PingUDP()
	default:
		fmt.Printf("Invalid protocol: %s\n", *protocol)
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Ping error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	pinger.PrintResults(results, *host, *packetSize, *protocol)
}
