package loginfail

import (
	"log"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_login_"

var (
	failuresDesc *prometheus.Desc
)

func init() {
	l := []string{"target"}
	failuresDesc = prometheus.NewDesc(prefix+"failures_total", "Total login failures", l, nil)
}

type loginfailCollector struct{}

func NewCollector() collector.RPCCollector {
	return &loginfailCollector{}
}

func (*loginfailCollector) Name() string {
	return "LoginFailures"
}

func (*loginfailCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- failuresDesc
}

func (c *loginfailCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show login failures")
	if err != nil {
		return err
	}
	count, err := c.Parse(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("Parse login failures for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}
	ch <- prometheus.MustNewConstMetric(failuresDesc, prometheus.CounterValue, float64(count), labelValues...)
	return nil
}