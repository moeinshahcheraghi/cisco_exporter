package qos

import (
    "regexp"
    "strings"
    "strconv"
    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_qos_"

var (
    dropsDesc *prometheus.Desc
)

func init() {
    dropsDesc = prometheus.NewDesc(prefix+"drops", "Number of QoS drops per interface and queue", []string{"target", "interface", "queue"}, nil)
}

type qosCollector struct{}

func NewCollector() collector.RPCCollector {
    return &qosCollector{}
}

func (*qosCollector) Name() string {
    return "QoS"
}

func (*qosCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- dropsDesc
}

func (c *qosCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
    out, err := client.RunCommand("show policy-map interface")
    if err != nil {
        return err
    }
    drops, err := parseQoSDrops(out)
    if err != nil {
        return err
    }
    for _, drop := range drops {
        labels := append(labelValues, drop.Interface, drop.Queue)
        ch <- prometheus.MustNewConstMetric(dropsDesc, prometheus.CounterValue, float64(drop.Count), labels...)
    }
    return nil
}

type Drop struct {
    Interface string
    Queue     string
    Count     int
}

func parseQoSDrops(output string) ([]Drop, error) {
    var drops []Drop
    lines := strings.Split(output, "\n")
    re := regexp.MustCompile(`(\S+)\s+queue\s+(\S+):\s+(\d+)\s+drops`)
    for _, line := range lines {
        matches := re.FindStringSubmatch(line)
        if matches != nil {
            count, _ := strconv.Atoi(matches[3])
            drops = append(drops, Drop{
                Interface: matches[1],
                Queue:     matches[2],
                Count:     count,
            })
        }
    }
    return drops, nil
}
