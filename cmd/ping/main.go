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

	// Parse command-line flags
	flag.Parse()

	// Create pinger with configured options
	p := pinger.NewPinger(*host,
		pinger.WithCount(*count),
		pinger.WithPacketSize(*packetSize),
		pinger.WithTimeout(*timeout),
		pinger.WithInterval(*interval),
	)

	// Perform ping
	results, err := p.Ping()
	if err != nil {
		fmt.Printf("Ping error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	pinger.PrintResults(results, *host, *packetSize)
}
