package pinger

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// PingConfig represents configuration for ping operations
type PingConfig struct {
	Host       string
	PacketSize int
	Count      int
	Timeout    time.Duration
	IntervalMS time.Duration
}

// PingResult stores the results of a ping attempt
type PingResult struct {
	SequenceNumber int
	RTT            time.Duration
	Success        bool
}

// Pinger handles the ping functionality
type Pinger struct {
	config PingConfig
}

// NewPinger creates a new Pinger with default or custom configuration
func NewPinger(host string, options ...func(*PingConfig)) *Pinger {
	config := PingConfig{
		Host:       host,
		PacketSize: 64, // Default packet size
		Count:      4,  // Default ping count
		Timeout:    2 * time.Second,
		IntervalMS: 1000 * time.Millisecond,
	}

	// Apply custom options
	for _, option := range options {
		option(&config)
	}

	return &Pinger{config: config}
}

// Option functions for configuring Pinger
func WithPacketSize(size int) func(*PingConfig) {
	return func(pc *PingConfig) {
		pc.PacketSize = size
	}
}

func WithCount(count int) func(*PingConfig) {
	return func(pc *PingConfig) {
		pc.Count = count
	}
}

func WithTimeout(timeout time.Duration) func(*PingConfig) {
	return func(pc *PingConfig) {
		pc.Timeout = timeout
	}
}

func WithInterval(intervalMS time.Duration) func(*PingConfig) {
	return func(pc *PingConfig) {
		pc.IntervalMS = intervalMS
	}
}

// Ping performs the actual ping operation
func (p *Pinger) Ping() ([]PingResult, error) {
	// Resolve the target host
	addr, err := net.ResolveIPAddr("ip4", p.config.Host)
	if err != nil {
		return nil, fmt.Errorf("resolution error: %v", err)
	}

	// Create ICMP connection
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("listen error: %v", err)
	}
	defer conn.Close()

	results := make([]PingResult, 0, p.config.Count)

	for seq := 1; seq <= p.config.Count; seq++ {
		// Create ICMP message
		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  seq,
				Data: make([]byte, p.config.PacketSize),
			},
		}

		// Populate data with sequence number
		for i := 0; i < p.config.PacketSize; i++ {
			msg.Body.(*icmp.Echo).Data[i] = byte(seq)
		}

		// Marshal the message
		wb, err := msg.Marshal(nil)
		if err != nil {
			return nil, fmt.Errorf("marshal error: %v", err)
		}

		// Send the ping
		start := time.Now()
		_, err = conn.WriteTo(wb, addr)
		if err != nil {
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
			})
			continue
		}

		// Read response
		rb := make([]byte, 1500)
		err = conn.SetReadDeadline(time.Now().Add(p.config.Timeout))
		if err != nil {
			return nil, fmt.Errorf("set deadline error: %v", err)
		}

		n, _, err := conn.ReadFrom(rb)
		if err != nil {
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
			})
			continue
		}

		// Parse response
		rm, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), rb[:n])
		if err != nil {
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
			})
			continue
		}

		// Validate response
		rtt := time.Since(start)
		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			results = append(results, PingResult{
				SequenceNumber: seq,
				RTT:            rtt,
				Success:        true,
			})
		default:
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
			})
		}

		// Wait between pings
		time.Sleep(p.config.IntervalMS)
	}

	return results, nil
}

// PrintResults displays ping results
func PrintResults(results []PingResult) {
	successCount := 0
	totalRTT := time.Duration(0)

	fmt.Println("Ping Results:")
	for _, result := range results {
		if result.Success {
			successCount++
			totalRTT += result.RTT
			fmt.Printf("Sequence %d: Success - RTT: %v\n", result.SequenceNumber, result.RTT)
		} else {
			fmt.Printf("Sequence %d: Failed\n", result.SequenceNumber)
		}
	}

	// Calculate statistics
	successRate := float64(successCount) / float64(len(results)) * 100
	var avgRTT time.Duration
	if successCount > 0 {
		avgRTT = totalRTT / time.Duration(successCount)
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("Total Attempts: %d\n", len(results))
	fmt.Printf("Successful Pings: %d\n", successCount)
	fmt.Printf("Success Rate: %.2f%%\n", successRate)
	if successCount > 0 {
		fmt.Printf("Average RTT: %v\n", avgRTT)
	}
}

// вывод udp icmp
// сравнение с ping
