package arp

import (
    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_arp_"

var (
    totalEntriesDesc *prometheus.Desc
)

func init() {
    l := []string{"target"}
    totalEntriesDesc = prometheus.NewDesc(prefix+"total_entries", "Total number of ARP entries", l, nil)
}

type arpCollector struct{}

func NewCollector() collector.RPCCollector {
    return &arpCollector{}
}

func (*arpCollector) Name() string {
    return "ARP"
}

func (*arpCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- totalEntriesDesc
}

func (c *arpCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
    out, err := client.RunCommand("show arp summary")
    if err != nil {
        return err
    }
    total, err := Parse(client.OSType, out)
    if err != nil {
        return err
    }
    ch <- prometheus.MustNewConstMetric(totalEntriesDesc, prometheus.GaugeValue, float64(total), labelValues...)
    return nil
}