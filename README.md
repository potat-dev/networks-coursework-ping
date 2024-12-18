# Ping Utility

A custom ping implementation in Golang with flexible configuration options, supporting both ICMP and UDP protocols.

## Features

-   Support for ICMP and UDP ping protocols
-   Customizable packet size
-   Configurable ping count
-   Adjustable timeout and interval
-   Detailed ping statistics
-   Output similar to standard ping utility
-   Uses common UDP ports (53, 123, 8053, 33434) to increase the chance of receiving a response
-   Sends a DNS query packet for UDP ping to elicit a response

## Prerequisites

-   Go 1.21 or higher
-   Root/administrator privileges for raw socket access

## Installation

1. Clone the repository
    ```bash
    git clone https://github.com/yourusername/ping-utility.git
    cd ping-utility
    ```
2. Install dependencies
    ```bash
    go mod tidy
    ```

## Usage

Run the ping utility with default settings (ICMP to Google's DNS):

```bash
sudo go run cmd/ping/main.go
```

Command-line options:

```
  -host string
        Host to ping (default "8.8.8.8")
  -count int
        Number of ping attempts (default 4)
  -size int
        Size of ping packet (default 64)
  -timeout duration
        Timeout for each ping (default 2s)
  -interval duration
        Interval between pings (default 1s)
  -protocol string
        Ping protocol (icmp or udp) (default "icmp")
  -port int
        Port to use for UDP ping (default 33434)
```

Examples:

1. ICMP Ping to Google's DNS:
    ```bash
    sudo go run cmd/ping/main.go -host 8.8.8.8
    ```
2. ICMP Ping with Custom Parameters:
    ```bash
    sudo go run cmd/ping/main.go -host google.com -count 5 -size 128 -timeout 3s
    ```
3. Basic UDP Ping:
    ```bash
    sudo go run cmd/ping/main.go -host 8.8.8.8 -protocol udp
    ```
4. UDP Ping with Custom Port:
    ```bash
    sudo go run cmd/ping/main.go -host 8.8.8.8 -protocol udp -port 53 -count 3
    ```
5. UDP Ping to a specific domain using DNS port:
    ```bash
    sudo go run cmd/ping/main.go -host google.com -protocol udp -port 53 -count 4
    ```

## Building

To build the executable:

```bash
make build
```

Then run with:

```bash
sudo ./build/ping [options]
```

## Makefile Targets

-   `build`: Builds the ping utility executable
-   `clean`: Cleans up the build directory
-   `test`: Runs tests
-   `deps`: Installs dependencies
-   `run`: Builds and runs the ping utility
-   `install`: Installs the ping utility to /usr/local/bin

## Limitations

-   Requires root/administrator privileges
-   Currently supports only IPv4
-   UDP ping relies on the target host responding to specific types of UDP packets (DNS queries on common ports)

## Troubleshooting

-   If UDP ping shows 100% packet loss, the target host may not be responding to the UDP packets. Try different ports or hosts.
-   Ensure you have the necessary privileges to run the utility (sudo or administrator).
-   Check your network firewall settings, as they might be blocking ICMP or UDP packets.

## Future Improvements

-   Add IPv6 support
-   Implement more detailed error messages
-   Add more protocol options (e.g., TCP ping)
-   Enhanced statistics and graphing
-   Configurable UDP packet types

## License

MIT License
