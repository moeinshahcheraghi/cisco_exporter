# Cisco Exporter

Cisco Exporter is a Prometheus exporter for monitoring Cisco network devices. It connects to Cisco switches and routers via SSH, executes various "show" commands, and exposes the collected data as Prometheus metrics.

## Supported Cisco OS

- IOS
- IOS XE
- NX-OS

## Installation

### Building from Source

To build Cisco Exporter from source, you need [Go](https://golang.org/) version 1.16 or later. Clone the repository and run:

```bash
go build -o cisco_exporter
```

Alternatively, you can download a pre-built binary from the [releases](https://github.com/moeinshahcheraghi/cisco_exporter/releases) page.

### Docker

You can run Cisco Exporter using Docker. First, build the Docker image:

```bash
docker build -t cisco_exporter .
```

Then, run the container, passing the necessary configuration. For example:

```bash
docker run -d -p 9362:9362 cisco_exporter -ssh.targets=192.168.1.1,192.168.1.2 -ssh.user=admin -ssh.password=admin_password
```

Alternatively, you can mount a configuration file:

```bash
docker run -d -p 9362:9362 -v /path/to/config.yaml:/config.yaml cisco_exporter -config.file=/config.yaml
```

### Kubernetes Helm

To deploy Cisco Exporter in Kubernetes using Helm, assuming there is a Helm chart in the repository under `./helm/cisco-exporter`, you can install it with:

```bash
helm install cisco-exporter ./helm/cisco-exporter --set ssh.targets="192.168.1.1,192.168.1.2" --set ssh.user="admin" --set ssh.password="admin_password"
```

For production deployments, consider using a values file to manage configuration, especially for sensitive data.

## Configuration

Cisco Exporter can be configured using a YAML file or command-line flags.

### Config File

Example `config.yaml`:

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
```

Run with:

```bash
./cisco_exporter -config.file=/path/to/config.yaml
```

### Command-Line Flags

Example:

```bash
./cisco_exporter -ssh.targets=192.168.1.1,192.168.1.2 -ssh.user=admin -ssh.password=admin_password
```

For a full list of flags, run:

```bash
./cisco_exporter -h
```

## Running the Exporter

By default, Cisco Exporter listens on `localhost:9362` and exposes metrics at `/metrics`. You can change the listen address and metrics path using flags.

## Prometheus Configuration

Add the following scrape config to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'cisco'
    static_configs:
      - targets: ['localhost:9362']
```

## Metrics

Cisco Exporter provides the following metrics:

| Metric Category   | Description                                                                                   |
|-------------------|-----------------------------------------------------------------------------------------------|
| **BGP**           | Monitors BGP session states (e.g., whether sessions are established) and statistics like the number of sent and received messages and received prefixes. |
| **Environment**   | Monitors environmental conditions via sensors (e.g., temperature) and the status of power supplies (e.g., whether they are functioning correctly).          |
| **Facts**         | Collects system-level information, such as the running OS version, CPU usage (over 5 seconds, 1 minute, and 5 minutes), and memory usage statistics (total, used, and free). |
| **Interfaces**    | Collects detailed network interface statistics, including input and output bytes, number of errors, drops, broadcast and multicast packets, and administrative and operational status. |
| **Optics**        | Monitors the performance of optical transceivers by tracking Tx (transmit) and Rx (receive) power levels.                  |
| **Stack Port**    | Checks the status of stack ports in stacked switch configurations, indicating whether each port is operational.         |

*Note: All metrics are prefixed with `cisco_` in Prometheus.*

## Dependencies

- Go 1.16 or later
- External libraries:
  - `golang.org/x/crypto/ssh`
  - `github.com/prometheus/client_golang`
  - `github.com/sirupsen/logrus`
  - `gopkg.in/yaml.v2`

## Contributing

Contributions are welcome! Please submit issues or pull requests on [GitHub](https://github.com/moeinshahcheraghi/cisco_exporter).

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Security Note

When configuring Cisco Exporter, especially in production environments, avoid passing sensitive information like passwords via command-line flags or plain text in configuration files. Instead, use SSH key-based authentication or manage secrets securely using tools like Kubernetes Secrets or environment variables.