package login

import (
    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_login_"

var (
    failuresTotalDesc *prometheus.Desc
)

func init() {
    l := []string{"target"}
    failuresTotalDesc = prometheus.NewDesc(prefix+"failures_total", "Total number of login failures", l, nil)
}

type loginCollector struct{}

func NewCollector() collector.RPCCollector {
    return &loginCollector{}
}

func (*loginCollector) Name() string {
    return "Login"
}

func (*loginCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- failuresTotalDesc
}

func (c *loginCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
    out, err := client.RunCommand("show login failures")
    if err != nil {
        return err
    }
    count, err := Parse(client.OSType, out)
    if err != nil {
        return err
    }
    ch <- prometheus.MustNewConstMetric(failuresTotalDesc, prometheus.CounterValue, float64(count), labelValues...)
    return nil
}