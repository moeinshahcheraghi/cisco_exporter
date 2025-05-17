package configlog

import (
	"log"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_config_"

var (
	changesDesc *prometheus.Desc
)

func init() {
	l := []string{"target"}
	changesDesc = prometheus.NewDesc(prefix+"changes_total", "Total configuration changes", l, nil)
}

type configlogCollector struct{}

func NewCollector() collector.RPCCollector {
	return &configlogCollector{}
}

func (*configlogCollector) Name() string {
	return "ConfigLog"
}

func (*configlogCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- changesDesc
}

func (c *configlogCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show logging | include Config")
	if err != nil {
		return err
	}
	count, err := c.Parse(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("Parse config log for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}
	ch <- prometheus.MustNewConstMetric(changesDesc, prometheus.CounterValue, float64(count), labelValues...)
	return nil
}