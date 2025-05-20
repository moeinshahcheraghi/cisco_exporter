package login

import (
    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_logging_"

var (
    loginSuccessDesc *prometheus.Desc
)

func init() {
    l := []string{"target"}
    loginSuccessDesc = prometheus.NewDesc(prefix+"login_success_total", "Total number of successful logins", l, nil)
}

type loggingCollector struct{}

func NewLoggingCollector() collector.RPCCollector {
    return &loggingCollector{}
}

func (*loggingCollector) Name() string {
    return "Logging"
}

func (*loggingCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- loginSuccessDesc
}

func (c *loggingCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
    out, err := client.RunCommand("show logging")
    if err != nil {
        return err
    }
    count, err := ParseLogging(client.OSType, out)
    if err != nil {
        return err
    }
    ch <- prometheus.MustNewConstMetric(loginSuccessDesc, prometheus.CounterValue, float64(count), labelValues...)
    return nil
}