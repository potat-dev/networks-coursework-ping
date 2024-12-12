# Ping Utility

A custom ping implementation in Golang with flexible configuration options.

## Features

-   Customizable packet size
-   Configurable ping count
-   Adjustable timeout and interval
-   Отображение типа пакета (ICMP)
-   Detailed ping statistics (packet loss, rtt min/avg/max)

## Prerequisites

-   Go 1.21 or higher
-   Root/administrator privileges for raw socket access

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

**Пример вывода:**

```
PING 8.8.8.8 (8.8.8.8) 64 bytes of data.
64 bytes from 8.8.8.8: icmp_seq=1 time=14.8 ms
64 bytes from 8.8.8.8: icmp_seq=2 time=14.6 ms
64 bytes from 8.8.8.8: icmp_seq=3 time=14.8 ms
64 bytes from 8.8.8.8: icmp_seq=4 time=14.6 ms
--- 8.8.8.8 ping statistics ---
4 packets transmitted, 4 received, 0.0% packet loss, time 59ms
rtt min/avg/max = 14.625ms/14.732ms/14.838ms
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

### Тип пакета

В текущей версии утилита поддерживает **только ICMP-пинг**. Поддержка UDP в настоящее время не реализована.

### Форматирование вывода

Вывод утилиты максимально приближен к выводу стандартной утилиты `ping`. Он включает в себя:

-   Заголовок с указанием целевого хоста, IP-адреса и размера пакета данных.
-   Для каждого успешного пинга: количество байт, IP-адрес, порядковый номер ICMP, время (time).
-   Статистику в конце: количество отправленных и полученных пакетов, процент потерь пакетов, общее время.
-   Статистику по времени приема-передачи (rtt): минимальное (min), среднее (avg), максимальное (max).

Example:

```bash
sudo go run cmd/ping/main.go -host google.com -count 5 -size 128 -timeout 3s
```

**Пример вывода:**

```
PING google.com (142.250.183.142) 128 bytes of data.
128 bytes from sof02s32-in-f14.1e100.net (142.250.183.142): icmp_seq=1 time=12.5 ms
128 bytes from sof02s32-in-f14.1e100.net (142.250.183.142): icmp_seq=2 time=12.6 ms
128 bytes from sof02s32-in-f14.1e100.net (142.250.183.142): icmp_seq=3 time=12.4 ms
128 bytes from sof02s32-in-f14.1e100.net (142.250.183.142): icmp_seq=4 time=12.5 ms
128 bytes from sof02s32-in-f14.1e100.net (142.250.183.142): icmp_seq=5 time=12.6 ms
--- google.com ping statistics ---
5 packets transmitted, 5 received, 0.0% packet loss, time 4006ms
rtt min/avg/max = 12.447ms/12.548ms/12.638ms
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

-   Requires root/administrator privileges
-   Currently supports only IPv4
-   Basic ping functionality
-   Поддерживается только ICMP-пинг
