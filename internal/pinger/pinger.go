package pinger

import (
	"encoding/binary"
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
	UDPPort    int
}

// PingResult stores the results of a ping attempt
type PingResult struct {
	SequenceNumber int
	RTT            time.Duration
	Success        bool
	From           string
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
		UDPPort:    33434, // Default UDP port
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

func WithUDPPort(udpPort int) func(*PingConfig) {
	return func(pc *PingConfig) {
		pc.UDPPort = udpPort
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

		n, from, err := conn.ReadFrom(rb)
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
				From:           from.String(),
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

// PingUDP performs the actual UDP ping operation
func (p *Pinger) PingUDP() ([]PingResult, error) {
	// Resolve the target host and port
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", p.config.Host, p.config.UDPPort))
	if err != nil {
		return nil, fmt.Errorf("UDP resolution error: %v", err)
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("UDP dial error: %v", err)
	}
	defer conn.Close()

	results := make([]PingResult, 0, p.config.Count)

	for seq := 1; seq <= p.config.Count; seq++ {
		// Prepare the data with process ID and sequence number
		pid := os.Getpid() & 0xffff
		data := make([]byte, p.config.PacketSize)
		binary.BigEndian.PutUint16(data[0:2], uint16(pid))
		binary.BigEndian.PutUint16(data[2:4], uint16(seq))

		// Send the ping
		start := time.Now()
		_, err = conn.Write(data)
		if err != nil {
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
			})
			continue
		}

		// Read response
		err = conn.SetReadDeadline(time.Now().Add(p.config.Timeout))
		if err != nil {
			return nil, fmt.Errorf("set UDP deadline error: %v", err)
		}

		buffer := make([]byte, 1500)
		_, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
			})
			continue
		}

		// Validate response
		rtt := time.Since(start)
		receivedPid := int(binary.BigEndian.Uint16(buffer[0:2]))
		receivedSeq := int(binary.BigEndian.Uint16(buffer[2:4]))

		if receivedPid == pid && receivedSeq == seq {
			results = append(results, PingResult{
				SequenceNumber: seq,
				RTT:            rtt,
				Success:        true,
				From:           from.String(),
			})
		} else {
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
func PrintResults(results []PingResult, host string, packetSize int, protocol string) {
	fmt.Printf("PING %s (%s) %d bytes of data using %s.\n", host, host, packetSize, protocol)

	successCount := 0
	var min, max, total time.Duration

	for _, result := range results {
		if result.Success {
			if protocol == "udp" {
				fmt.Printf("%d bytes from %s: udp_seq=%d time=%v\n",
					packetSize, result.From, result.SequenceNumber, result.RTT)
			} else {
				fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
					packetSize, result.From, result.SequenceNumber, result.RTT)
			}

			successCount++

			if min == 0 || result.RTT < min {
				min = result.RTT
			}
			if max == 0 || result.RTT > max {
				max = result.RTT
			}
			total += result.RTT
		} else {
			if protocol == "udp" {
				fmt.Printf("Request timeout for udp_seq %d\n", result.SequenceNumber)
			} else {
				fmt.Printf("Request timeout for icmp_seq %d\n", result.SequenceNumber)
			}
		}
	}

	// Calculate statistics
	packetLoss := float64(len(results)-successCount) / float64(len(results)) * 100
	var avg time.Duration
	if successCount > 0 {
		avg = total / time.Duration(successCount)
	}

	fmt.Printf("--- %s ping statistics ---\n", host)
	fmt.Printf("%d packets transmitted, %d received, %.1f%% packet loss, time %dms\n",
		len(results), successCount, packetLoss, total.Milliseconds())

	if successCount > 0 {
		fmt.Printf("rtt min/avg/max = %v/%v/%v\n", min, avg, max)
	}
}
