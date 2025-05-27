

# Cisco Exporter

**Cisco Exporter** is a Prometheus exporter designed to monitor Cisco network devices by collecting metrics via SSH. It executes various Cisco CLI `show` commands, parses the output, and exposes the data as Prometheus metrics. The tool features a modular architecture, making it extensible and suitable for monitoring a wide range of Cisco devices.

## Supported Cisco OS
- **IOS**
- **IOS XE**
- **NX-OS**

## Table of Contents
- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Exporter](#running-the-exporter)
- [Prometheus Configuration](#prometheus-configuration)
- [Metrics](#metrics)
- [Dependencies](#dependencies)
- [Contributing](#contributing)
- [License](#license)
- [Security Note](#security-note)

## Features
Cisco Exporter offers a range of features to monitor Cisco network devices effectively:

- **Multi-OS Support**: Compatible with Cisco IOS, IOS XE, and NX-OS.
- **Modular Collectors**: Includes collectors for BGP, environment, system facts, interfaces, optics, stack ports, and routing tables (ARP, MAC, IPv4, IPv6).
- **Flexible Configuration**: Supports YAML configuration files and command-line flags with per-device overrides.
- **Caching Mechanism**: Reduces redundant SSH commands with a simple in-memory cache.
- **Debugging and Logging**: Detailed logging with a debug mode for troubleshooting.
- **Secure Connections**: Supports SSH key-based authentication and legacy ciphers for older devices.
- **Extended Table Metrics (New)**: Collects ARP table entries, MAC address table counts, and IPv4/IPv6 routing table sizes.
- **Enhanced Optics Collection (New)**: Supports per-interface transceiver data for IOS XE and batch collection for IOS/NX-OS.
- **Stack Port Monitoring (New)**: Monitors the status of stack ports in stacked switches.
- **Robust Error Handling (New)**: Graceful handling of SSH timeouts and command failures with detailed debug logs.
- **Performance Optimization (New)**: Batch size configuration for SSH responses to efficiently handle large command outputs.

## Architecture
Cisco Exporter follows a modular and layered architecture to ensure scalability and maintainability:

- **Connector Layer** (`./connector/`):
  - Manages SSH connections to Cisco devices using `golang.org/x/crypto/ssh`.
  - Supports authentication via passwords or SSH keys, with options for legacy ciphers and timeouts.
- **RPC Layer** (`./rpc/`):
  - Abstracts command execution and OS identification.
  - Implements a caching mechanism to store command outputs, reducing SSH overhead.
  - Identifies the device OS (IOS, IOS XE, NX-OS) using `show version`.
- **Collector Layer** (`./bgp/`, `./environment/`, etc.):
  - Each collector (e.g., BGP, Interfaces) implements the `RPCCollector` interface.
  - Executes specific CLI commands, parses outputs with regex, and exposes metrics to Prometheus.
  - Collectors are independent, allowing easy addition of new metric types.
- **Main Application** (`./main.go`, `./cisco_collector.go`):
  - Orchestrates device connections, collector registration, and metric exposure via an HTTP endpoint.
  - Uses Go concurrency (`sync.WaitGroup`) to collect metrics from multiple devices in parallel.

This separation enables developers to extend functionality by adding new collectors without modifying the core logic.

## Installation

### Building from Source
Requires Go 1.16 or later:

```bash
git clone https://github.com/moeinshahcheraghi/cisco_exporter.git
cd cisco_exporter
go build -o cisco_exporter
```

Alternatively, download a pre-built binary from the [releases page](https://github.com/moeinshahcheraghi/cisco_exporter/releases).

### Docker
Build and run with Docker:

```bash
docker build -t cisco_exporter .
docker run -d -p 9362:9362 cisco_exporter -ssh.targets=192.168.1.1,192.168.1.2 -ssh.user=admin -ssh.password=admin_password
```

Mount a config file if preferred:

```bash
docker run -d -p 9362:9362 -v /path/to/config.yaml:/config.yaml cisco_exporter -config.file=/config.yaml
```

### Kubernetes Helm
Deploy using Helm (assuming a chart exists in `./helm/cisco-exporter`):

```bash
helm install cisco-exporter ./helm/cisco-exporter --set ssh.targets="192.168.1.1,192.168.1.2" --set ssh.user="admin" --set ssh.password="admin_password"
```

Use a `values.yaml` file for production to manage sensitive data securely.

## Configuration
Cisco Exporter supports both YAML configuration files and command-line flags.

### Config File Example
```yaml
debug: false
legacy_ciphers: false
timeout: 5
batch_size: 10000
username: cisco_exporter
password: your_password
key_file: /path/to/keyfile
devices:
  - host: 192.168.1.1
    username: admin
    password: admin_password
features:
  bgp: true
  environment: true
  facts: true
  interfaces: true
  optics: true
  stack_port: true
  tables_arp: true
  tables_mac: true
  tables_route_ipv4: true
  tables_route_ipv6: true
```

Run with:

```bash
./cisco_exporter -config.file=/path/to/config.yaml
```

### Command-Line Flags
Example:

```bash
./cisco_exporter -ssh.targets=192.168.1.1,192.168.1.2 -ssh.user=admin -ssh.password=admin_password -bgp.enabled=true -tables_arp.enabled=true
```

View all options:

```bash
./cisco_exporter -h
```

## Running the Exporter
By default, the exporter listens on `localhost:9362` and exposes metrics at `/metrics`. Customize with:

```bash
./cisco_exporter -web.listen-address=":9362" -web.telemetry-path="/metrics"
```

## Prometheus Configuration
Add to `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'cisco'
    static_configs:
      - targets: ['localhost:9362']
```

## Metrics
| **Category**    | **Description**                                                                 |
|------------------|---------------------------------------------------------------------------------|
| **BGP**         | Monitors BGP session states (1 = Established), received prefixes, and messages. |
| **Environment** | Tracks sensor temperatures and power supply status (1 = OK, 0 = Not OK).       |
| **Facts**       | Collects OS version, CPU usage (5s, 1m, 5m, interrupts), and memory stats.      |
| **Interfaces**  | Monitors traffic (bytes), errors, drops, broadcasts, multicasts, and status.    |
| **Optics**      | Tracks optical transceiver Tx/Rx power levels.                                  |
| **Stack Port**  | Monitors stack port status in stacked switches (1 = OK, 0 = Not OK).            |
| **Tables**      | Counts ARP entries, MAC addresses, and IPv4/IPv6 routes.                        |

Metrics are prefixed with `cisco_`.

## Dependencies
- **Go**: 1.16+
- **External Libraries**:
  - `golang.org/x/crypto/ssh`
  - `github.com/prometheus/client_golang`
  - `github.com/sirupsen/logrus`
  - `gopkg.in/yaml.v2`

## Contributing
Contributions are welcome! Please submit issues or pull requests on [GitHub](https://github.com/moeinshahcheraghi/cisco_exporter). To add a new collector:
1. Create a new package under `./` (e.g., `./qos/`).
2. Implement the `RPCCollector` interface.
3. Register it in `collectors.go`.



## Security Note
Avoid passing sensitive data (e.g., passwords) via command-line flags in production. Use SSH keys or secure secret management tools (e.g., Kubernetes Secrets, environment variables) instead.

---
