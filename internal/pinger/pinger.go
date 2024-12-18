package pinger

import (
	"fmt"
	"net"
	"os"
	// "strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// PingProtocol defines the type of ping protocol
type PingProtocol string

const (
	ProtocolICMP PingProtocol = "ICMP"
	ProtocolUDP  PingProtocol = "UDP"
)

// PingConfig represents configuration for ping operations
type PingConfig struct {
	Host           string
	PacketSize     int
	Count          int
	Timeout        time.Duration
	IntervalMS     time.Duration
	Protocol       PingProtocol
	Port           int
}

// PingResult stores the results of a ping attempt
type PingResult struct {
	SequenceNumber int
	RTT            time.Duration
	Success        bool
	Protocol       PingProtocol
}

// Pinger handles the ping functionality
type Pinger struct {
	config PingConfig
}

// NewPinger creates a new Pinger with default or custom configuration
func NewPinger(host string, options ...func(*PingConfig)) *Pinger {
	config := PingConfig{
		Host:           host,
		PacketSize:     64,    // Default packet size
		Count:          4,     // Default ping count
		Timeout:        2 * time.Second,
		IntervalMS:     1000 * time.Millisecond,
		Protocol:       ProtocolICMP,
		Port:           33434, // Default traceroute-like port for UDP
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

func WithProtocol(protocol PingProtocol) func(*PingConfig) {
	return func(pc *PingConfig) {
		pc.Protocol = protocol
	}
}

func WithPort(port int) func(*PingConfig) {
	return func(pc *PingConfig) {
		pc.Port = port
	}
}

// Ping performs the actual ping operation
func (p *Pinger) Ping() ([]PingResult, error) {
	switch p.config.Protocol {
	case ProtocolICMP:
		return p.pingICMP()
	case ProtocolUDP:
		return p.pingUDP()
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", p.config.Protocol)
	}
}

// pingICMP performs ICMP ping
func (p *Pinger) pingICMP() ([]PingResult, error) {
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
				Protocol:       ProtocolICMP,
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
				Protocol:       ProtocolICMP,
			})
			continue
		}

		// Parse response
		rm, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), rb[:n])
		if err != nil {
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
				Protocol:       ProtocolICMP,
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
				Protocol:       ProtocolICMP,
			})
		default:
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
				Protocol:       ProtocolICMP,
			})
		}

		// Wait between pings
		time.Sleep(p.config.IntervalMS)
	}

	return results, nil
}

// pingUDP performs UDP ping
func (p *Pinger) pingUDP() ([]PingResult, error) {
	// Resolve the target host
	raddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", p.config.Host, p.config.Port))
	if err != nil {
		return nil, fmt.Errorf("resolution error: %v", err)
	}

	results := make([]PingResult, 0, p.config.Count)

	for seq := 1; seq <= p.config.Count; seq++ {
		// Create UDP connection
		conn, err := net.DialTimeout("udp4", raddr.String(), p.config.Timeout)
		if err != nil {
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
				Protocol:       ProtocolUDP,
			})
			continue
		}
		defer conn.Close()

		// Prepare data
		data := make([]byte, p.config.PacketSize)
		for i := 0; i < p.config.PacketSize; i++ {
			data[i] = byte(seq)
		}

		// Send ping
		start := time.Now()
		_, err = conn.Write(data)
		if err != nil {
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
				Protocol:       ProtocolUDP,
			})
			continue
		}

		// Set read deadline
		err = conn.SetReadDeadline(time.Now().Add(p.config.Timeout))
		if err != nil {
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
				Protocol:       ProtocolUDP,
			})
			continue
		}

		// Read response
		rb := make([]byte, 1500)
		_, err = conn.Read(rb)
		if err != nil {
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
				Protocol:       ProtocolUDP,
			})
			continue
		}

		// Calculate RTT
		rtt := time.Since(start)

		// Successful ping
		results = append(results, PingResult{
			SequenceNumber: seq,
			RTT:            rtt,
			Success:        true,
			Protocol:       ProtocolUDP,
		})

		// Wait between pings
		time.Sleep(p.config.IntervalMS)
	}

	return results, nil
}

// PrintResults displays ping results in a format similar to standard ping utility
func PrintResults(target string, results []PingResult) {
	if len(results) == 0 {
		fmt.Printf("No ping attempts made to %s\n", target)
		return
	}

	// Print header
	fmt.Printf("Pinging %s using %s protocol:\n\n", target, results[0].Protocol)

	successCount := 0
	totalRTT := time.Duration(0)
	minRTT := time.Duration(0)
	maxRTT := time.Duration(0)
	rtts := []time.Duration{}

	// Compute statistics
	for _, result := range results {
		if result.Success {
			successCount++
			totalRTT += result.RTT
			rtts = append(rtts, result.RTT)

			if minRTT == 0 || result.RTT < minRTT {
				minRTT = result.RTT
			}
			if result.RTT > maxRTT {
				maxRTT = result.RTT
			}

			fmt.Printf("Reply from %s: bytes=%d time=%v\n", 
				target, len(result.Protocol), result.RTT)
		} else {
			fmt.Printf("Request timed out.\n")
		}
	}

	// Calculate statistics
	successRate := float64(successCount) / float64(len(results)) * 100
	var avgRTT time.Duration
	// var mdevRTT time.Duration

	if successCount > 0 {
		avgRTT = totalRTT / time.Duration(successCount)

		// Calculate mean deviation (mdev)
		var sumSquareDiff time.Duration
		for _, rtt := range rtts {
			diff := rtt - avgRTT
			sumSquareDiff += diff * diff
		}
		// mdevRTT = time.Duration(float64(sumSquareDiff) / float64(successCount))
	}

	// Print summary
	fmt.Printf("\nPing statistics for %s:\n", target)
	fmt.Printf("\tPackets: Sent = %d, Received = %d, Lost = %d (%0.2f%% loss)\n", 
		len(results), successCount, len(results) - successCount, 100 - successRate)

	if successCount > 0 {
		fmt.Printf("Approximate round trip times in milli-seconds:\n")
		fmt.Printf("\tMinimum = %vms, Maximum = %vms, Average = %vms\n", 
			minRTT.Milliseconds(), maxRTT.Milliseconds(), avgRTT.Milliseconds())
	}
}