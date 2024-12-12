# Ping Utility

A custom ping implementation in Golang with flexible configuration options.

## Features

- Customizable packet size
- Configurable ping count
- Adjustable timeout and interval
- Detailed ping statistics

## Prerequisites

- Go 1.21 or higher
- Root/administrator privileges for raw socket access

## Installation

1. Clone the repository
```bash
git clone https://github.com/potat-dev/ping-utility.git
cd ping-utility
```

2. Install dependencies
```bash
go mod tidy
```

## Usage

Run the ping utility with default settings:
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
```

Example:
```bash
sudo go run cmd/ping/main.go -host google.com -count 5 -size 128 -timeout 3s
```

## Building

To build the executable:
```bash
go build -o ping cmd/ping/main.go
```

Then run with:
```bash
sudo ./ping
```

## Limitations

- Requires root/administrator privileges
- Currently supports only IPv4
- Basic ping functionality
