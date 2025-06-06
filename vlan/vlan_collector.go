package vlan

import (
    "regexp"
    "strings"

    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_vlan_"

var (
    countDesc *prometheus.Desc
)

func init() {
    countDesc = prometheus.NewDesc(prefix+"count", "Total number of VLANs", []string{"target"}, nil)
}

type vlanCollector struct{}

func NewCollector() collector.RPCCollector {
    return &vlanCollector{}
}

func (*vlanCollector) Name() string {
    return "VLAN"
}

func (*vlanCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- countDesc
}

func (c *vlanCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
    out, err := client.RunCommand("show vlan brief")
    if err != nil {
        return err
    }
    count, err := parseVLANCount(out)
    if err != nil {
        return err
    }
    ch <- prometheus.MustNewConstMetric(countDesc, prometheus.GaugeValue, float64(count), labelValues...)
    return nil
}

func parseVLANCount(output string) (int, error) {
    lines := strings.Split(output, "\n")
    count := 0
    re := regexp.MustCompile(`^\d+\s+`)
    for _, line := range lines {
        if re.MatchString(line) {
            count++
        }
    }
    return count, nil
}
