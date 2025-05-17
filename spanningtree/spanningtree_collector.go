package spanningtree

import (
	"log"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_spanning_tree_"

var (
	portStateDesc *prometheus.Desc
)

func init() {
	l := []string{"target", "vlan", "interface"}
	portStateDesc = prometheus.NewDesc(prefix+"port_state", "Port state (1 = Forwarding, 0 = Blocking, etc.)", l, nil)
}

type spanningtreeCollector struct{}

func NewCollector() collector.RPCCollector {
	return &spanningtreeCollector{}
}

func (*spanningtreeCollector) Name() string {
	return "SpanningTree"
}

func (*spanningtreeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- portStateDesc
}

func (c *spanningtreeCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show spanning-tree detail")
	if err != nil {
		return err
	}
	items, err := c.Parse(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("Parse spanning tree for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}
	for _, item := range items {
		l := append(labelValues, item.VLAN, item.Interface)
		ch <- prometheus.MustNewConstMetric(portStateDesc, prometheus.GaugeValue, float64(item.State), l...)
	}
	return nil
}