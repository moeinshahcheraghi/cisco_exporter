package stackport

import (
	"log"

	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_stack_port_"

var (
	stackPortStatusDesc *prometheus.Desc
)

func init() {
	labels := []string{"target", "switch", "port"}
	stackPortStatusDesc = prometheus.NewDesc(prefix+"status", "Status of stack ports (1 OK, 0 Not OK)", labels, nil)
}

type stackPortCollector struct{}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &stackPortCollector{}
}

// Name returns the name of the collector
func (*stackPortCollector) Name() string {
	return "StackPort"
}

// Describe describes the metrics
func (*stackPortCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- stackPortStatusDesc
}

// Collect collects metrics from Cisco
func (c *stackPortCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show switch stack-ports")
	if err != nil {
		return err
	}
	items, err := Parse(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("Parse stack ports for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}

	for _, item := range items {
		l := append(labelValues, item.Switch, item.Port)
		val := 0.0
		if item.OK {
			val = 1.0
		}
		ch <- prometheus.MustNewConstMetric(stackPortStatusDesc, prometheus.GaugeValue, val, l...)
	}

	return nil
}