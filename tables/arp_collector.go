package tables

import (
	"regexp"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/moeinshahcheraghi/cisco_exporter/util"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_tables_"

var (
	arpEntriesDesc = prometheus.NewDesc(prefix+"arp_entries", "Number of ARP entries", []string{"target"}, nil)
)

type arpCollector struct{}

func NewARPCollector() collector.RPCCollector {
	return &arpCollector{}
}

func (*arpCollector) Name() string {
	return "TablesARP"
}

func (*arpCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- arpEntriesDesc
}

func (c *arpCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show ip arp summary")
	if err != nil {
		return err
	}
	arpCount := parseARP(out)
	ch <- prometheus.MustNewConstMetric(arpEntriesDesc, prometheus.GaugeValue, arpCount, labelValues...)
	return nil
}

func parseARP(output string) float64 {
	re := regexp.MustCompile(`Total number of entries:\s*(\d+)`)
	matches := re.FindStringSubmatch(output)
	if matches != nil {
		return util.Str2float64(matches[1])
	}
	return 0
}