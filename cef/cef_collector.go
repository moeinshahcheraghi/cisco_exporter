package cef

import (
	"log"
	"regexp"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_cef_"

var (
	dropsDesc *prometheus.Desc
)

func init() {
	l := []string{"target", "interface"}
	dropsDesc = prometheus.NewDesc(prefix+"drops_total", "CEF packet drops", l, nil)
}

type cefCollector struct{}

func NewCollector() collector.RPCCollector {
	return &cefCollector{}
}

func (*cefCollector) Name() string {
	return "CEF"
}

func (*cefCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- dropsDesc
}

func (c *cefCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show interfaces stats | exclude disabled")
	if err != nil {
		return err
	}
	interfaces, err := c.ParseInterfaces(client.OSType, out)
	if err != nil {
		return err
	}
	for _, iface := range interfaces {
		out, err := client.RunCommand("show cef interface " + iface)
		if err != nil {
			continue
		}
		drops, err := c.Parse(client.OSType, out)
		if err != nil {
			continue
		}
		l := append(labelValues, iface)
		ch <- prometheus.MustNewConstMetric(dropsDesc, prometheus.CounterValue, float64(drops), l...)
	}
	return nil
}