package arp

import (
	"log"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_arp_"

var (
	entriesDesc *prometheus.Desc
)

func init() {
	l := []string{"target"}
	entriesDesc = prometheus.NewDesc(prefix+"entries_total", "Total ARP entries", l, nil)
}

type arpCollector struct{}

func NewCollector() collector.RPCCollector {
	return &arpCollector{}
}

func (*arpCollector) Name() string {
	return "ARP"
}

func (*arpCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- entriesDesc
}

func (c *arpCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show arp summary")
	if err != nil {
		return err
	}
	entries, err := c.Parse(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("Parse ARP for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}
	ch <- prometheus.MustNewConstMetric(entriesDesc, prometheus.GaugeValue, float64(entries), labelValues...)
	return nil
}