package pinger

import (
	"fmt"
	"net"
	"os"
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
	Host       string
	PacketSize int
	Count      int
	Timeout    time.Duration
	Interval   time.Duration
	Protocol   PingProtocol
	Port       int
}

// PingResult stores the results of a ping attempt
type PingResult struct {
	SequenceNumber  int
	RTT             time.Duration
	Success         bool
	Protocol        PingProtocol
	DataSentSize    int
	DataReceiveSize int
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
		Interval:   1000 * time.Millisecond,
		Protocol:   ProtocolICMP,
		Port:       33434, // Default traceroute-like port for UDP
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

func WithInterval(interval time.Duration) func(*PingConfig) {
	return func(pc *PingConfig) {
		pc.Interval = interval
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
	// Print header before starting the ping process
	fmt.Printf("Pinging %s using %s protocol:\n\n", p.config.Host, p.config.Protocol)

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

		// Calculate the actual size of data sent
		dataSentSize := 0
		if seq <= 255 {
			dataSentSize = 1
		} else if seq <= 65535 {
			dataSentSize = 2
		} else if seq <= 16777215 {
			dataSentSize = 3
		} else {
			dataSentSize = 4
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
			fmt.Printf("Request timed out.\n")
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
			fmt.Printf("Request timed out.\n")
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
			fmt.Printf("Request timed out.\n")
			continue
		}

		// Validate response
		rtt := time.Since(start)
		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			results = append(results, PingResult{
				SequenceNumber:  seq,
				RTT:             rtt,
				Success:         true,
				Protocol:        ProtocolICMP,
				DataSentSize:    dataSentSize, // Record data sent size
				DataReceiveSize: n,            // Record data received size
			})
			fmt.Printf("Reply from %s: bytes=%d time=%v\n",
				p.config.Host, n, rtt)
		default:
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
				Protocol:       ProtocolICMP,
			})
			fmt.Printf("Request timed out.\n")
		}

		// Wait between pings
		time.Sleep(p.config.Interval)
	}

	return results, nil
}

// pingUDP performs UDP ping
func (p *Pinger) pingUDP() ([]PingResult, error) {
	// Use common DNS ports to increase chance of response
	commonPorts := []int{53, 123, 8053, 33434}

	results := make([]PingResult, 0, p.config.Count)

	for seq := 1; seq <= p.config.Count; seq++ {
		success := false
		for _, port := range commonPorts {
			// Resolve the target host
			raddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", p.config.Host, port))
			if err != nil {
				continue
			}

			// Create UDP connection
			conn, err := net.DialTimeout("udp4", raddr.String(), p.config.Timeout)
			if err != nil {
				continue
			}

			// Prepare data (DNS query for Google's DNS)
			data := []byte{
				0x00, 0x01, // Transaction ID
				0x01, 0x00, // Flags (standard query)
				0x00, 0x01, // Questions: 1
				0x00, 0x00, // Answer RRs: 0
				0x00, 0x00, // Authority RRs: 0
				0x00, 0x00, // Additional RRs: 0
				0x07, 'g', 'o', 'o', 'g', 'l', 'e', 0x03, 'c', 'o', 'm', 0x00, // Domain name
				0x00, 0x01, // Type: A (IPv4 address)
				0x00, 0x01, // Class: IN
			}

			// Calculate the size of the DNS query
			dataSentSize := len(data)

			// Send ping
			start := time.Now()
			_, err = conn.Write(data)
			if err != nil {
				conn.Close()
				continue
			}

			// Set read deadline
			err = conn.SetReadDeadline(time.Now().Add(p.config.Timeout))
			if err != nil {
				conn.Close()
				continue
			}

			// Read response
			rb := make([]byte, 1500)
			n, err := conn.Read(rb) // Read data into rb and get the number of bytes read

			// Close connection immediately after read
			conn.Close()

			if err != nil {
				fmt.Printf("Request timed out.\n")
				continue
			}

			// Calculate RTT
			rtt := time.Since(start)

			// Successful ping
			results = append(results, PingResult{
				SequenceNumber:  seq,
				RTT:             rtt,
				Success:         true,
				Protocol:        ProtocolUDP,
				DataSentSize:    dataSentSize, // Record data sent size
				DataReceiveSize: n,            // Record data received size
			})
			fmt.Printf("Reply from %s: bytes=%d time=%v\n",
				p.config.Host, n, rtt)

			success = true
			break // Break out of the port loop if successful
		}

		if !success {
			results = append(results, PingResult{
				SequenceNumber: seq,
				Success:        false,
				Protocol:       ProtocolUDP,
			})
			fmt.Printf("Request timed out.\n")
		}

		// Wait between pings
		time.Sleep(p.config.Interval)
	}

	return results, nil
}

// PrintResults displays ping results in a format similar to standard ping utility
func PrintResults(target string, results []PingResult) {
	successCount := 0
	totalRTT := time.Duration(0)
	minRTT := time.Duration(0)
	maxRTT := time.Duration(0)
	rtts := []time.Duration{}

	// Initialize total sent and received sizes
	totalDataSent := 0
	totalDataReceived := 0

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

			// Accumulate sent and received sizes
			totalDataSent += result.DataSentSize
			totalDataReceived += result.DataReceiveSize
		}
	}

	// Calculate statistics
	successRate := float64(successCount) / float64(len(results)) * 100
	var avgRTT time.Duration

	if successCount > 0 {
		avgRTT = totalRTT / time.Duration(successCount)

		// Calculate mean deviation (mdev)
		var sumSquareDiff time.Duration
		for _, rtt := range rtts {
			diff := rtt - avgRTT
			sumSquareDiff += diff * diff
		}
	}

	// Print summary
	fmt.Printf("\nPing statistics for %s:\n", target)
	fmt.Printf("\tPackets: Sent = %d, Received = %d, Lost = %d (%0.2f%% loss)\n",
		len(results), successCount, len(results)-successCount, 100-successRate)

	if successCount > 0 {
		fmt.Printf("Approximate round trip times in milli-seconds:\n")
		fmt.Printf("\tMinimum = %vms, Maximum = %vms, Average = %vms\n",
			minRTT.Milliseconds(), maxRTT.Milliseconds(), avgRTT.Milliseconds())

		// Print total data sent and received
		fmt.Printf("Total data sent: %d bytes, Total data received: %d bytes\n", totalDataSent, totalDataReceived)
	}
}
