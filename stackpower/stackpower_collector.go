package stackports

import (
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"log"
)

var (
	stackPortStatusDesc = prometheus.NewDesc("cisco_stackports_status", "Stack port status (1=OK, 0=Not OK)", []string{"target", "switch", "port"}, nil)
)

type stackPortsCollector struct{}

func NewCollector() collector.RPCCollector {
	return &stackPortsCollector{}
}

func (c *stackPortsCollector) Name() string {
	return "StackPorts"
}

func (c *stackPortsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- stackPortStatusDesc
}

func (c *stackPortsCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labels []string) error {
	out, err := client.RunCommand("show switch stack-ports")
	if err != nil {
		log.Println("Failed to run stack-port command:", err)
		return err
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] != "Switch#" {
			switchNum := fields[0]
			for i, status := range fields[1:] {
				port := "Port" + strconv.Itoa(i+1)
				val := 0.0
				if status == "OK" {
					val = 1.0
				}
				ch <- prometheus.MustNewConstMetric(stackPortStatusDesc, prometheus.GaugeValue, val, labels[0], switchNum, port)
			}
		}
	}

	return nil
}
