package stack

import (
	"context"
	"regexp"

	"github.com/moeinshahcheraghi/cisco_exporter/connector"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	stackPortsCommand = "show switch stack-ports"
)

// NewCollector creates a new collector
func NewCollector() *Collector {
	return &Collector{}
}

// Collector collects interface metrics
type Collector struct {
}

// Name of the collector
func (*Collector) Name() string {
	return "stack"
}

// Collect metrics
func (c *Collector) Collect(ctx context.Context, device *connector.Device, ch chan<- prometheus.Metric) error {
	cmd, err := device.ExecCommand(stackPortsCommand)
	if err != nil {
		return err
	}

	metrics, err := ParseStackPorts(device.Hostname, cmd)
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		ch <- metric
	}

	return nil
}
