package spanningtree

import (
    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_spanningtree_"

var (
    blockedPortsDesc *prometheus.Desc
)

func init() {
    l := []string{"target", "instance"}
    blockedPortsDesc = prometheus.NewDesc(prefix+"blocked_ports_total", "Total number of blocked ports per instance", l, nil)
}

type spanningtreeCollector struct{}

func NewCollector() collector.RPCCollector {
    return &spanningtreeCollector{}
}

func (*spanningtreeCollector) Name() string {
    return "SpanningTree"
}

func (*spanningtreeCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- blockedPortsDesc
}

func (c *spanningtreeCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
    out, err := client.RunCommand("show spanning-tree detail")
    if err != nil {
        return err
    }
    instances, err := Parse(client.OSType, out)
    if err != nil {
        return err
    }
    for _, instance := range instances {
        l := append(labelValues, instance.InstanceID)
        ch <- prometheus.MustNewConstMetric(blockedPortsDesc, prometheus.GaugeValue, float64(instance.BlockedPorts), l...)
    }
    return nil
}