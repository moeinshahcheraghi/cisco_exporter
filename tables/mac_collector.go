package tables

import (
	"regexp"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/moeinshahcheraghi/cisco_exporter/util"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	macAddressesDesc = prometheus.NewDesc(prefix+"mac_addresses", "Number of MAC addresses", []string{"target"}, nil)
)

type macCollector struct{}

func NewMACCollector() collector.RPCCollector {
	return &macCollector{}
}

func (*macCollector) Name() string {
	return "TablesMAC"
}

func (*macCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- macAddressesDesc
}

func (c *macCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show mac address-table count")
	if err != nil {
		return err
	}
	macCount := parseMAC(out)
	ch <- prometheus.MustNewConstMetric(macAddressesDesc, prometheus.GaugeValue, macCount, labelValues...)
	return nil
}

func parseMAC(output string) float64 {
	re := regexp.MustCompile(`Total Mac Addresses\s+:\s*(\d+)`)
	matches := re.FindAllStringSubmatch(output, -1)
	total := 0.0
	for _, match := range matches {
		total += util.Str2float64(match[1])
	}
	return total
}