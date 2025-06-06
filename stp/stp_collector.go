package stp

import (
    "errors"
    "regexp"

    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_stp_"

var (
    instancesDesc *prometheus.Desc
)

func init() {
    instancesDesc = prometheus.NewDesc(prefix+"instances", "Number of STP instances", []string{"target"}, nil)
}

type stpCollector struct{}

func NewCollector() collector.RPCCollector {
    return &stpCollector{}
}

func (*stpCollector) Name() string {
    return "STP"
}

func (*stpCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- instancesDesc
}

func (c *stpCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
    out, err := client.RunCommand("show spanning-tree summary")
    if err != nil {
        return err
    }
    instances, err := parseSTPInstances(out)
    if err != nil {
        return err
    }
    ch <- prometheus.MustNewConstMetric(instancesDesc, prometheus.GaugeValue, float64(instances), labelValues...)
    return nil
}

func parseSTPInstances(output string) (int, error) {
    re := regexp.MustCompile(`Total Instances: (\d+)`)
    matches := re.FindStringSubmatch(output)
    if matches == nil {
        return 0, errors.New("STP instances not found")
    }
    count, _ := strconv.Atoi(matches[1])
    return count, nil
}
