package uptime

import (
    "errors"
    "regexp"
    "strings"

    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_uptime_"

var (
    uptimeDesc *prometheus.Desc
)

func init() {
    uptimeDesc = prometheus.NewDesc(prefix+"seconds", "Device uptime in seconds", []string{"target"}, nil)
}

type uptimeCollector struct{}

func NewCollector() collector.RPCCollector {
    return &uptimeCollector{}
}

func (*uptimeCollector) Name() string {
    return "Uptime"
}

func (*uptimeCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- uptimeDesc
}

func (c *uptimeCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
    out, err := client.RunCommand("show version")
    if err != nil {
        return err
    }
    uptimeSeconds, err := parseUptime(out)
    if err != nil {
        return err
    }
    ch <- prometheus.MustNewConstMetric(uptimeDesc, prometheus.GaugeValue, uptimeSeconds, labelValues...)
    return nil
}

func parseUptime(output string) (float64, error) {
    re := regexp.MustCompile(`uptime is ([\w\s,]+)`)
    matches := re.FindStringSubmatch(output)
    if matches == nil {
        return 0, errors.New("uptime not found in output")
    }
    return convertUptimeToSeconds(matches[1]), nil
}

func convertUptimeToSeconds(uptimeStr string) float64 {
    var totalSeconds float64
    parts := strings.Split(uptimeStr, ", ")
    for _, part := range parts {
        fields := strings.Fields(part)
        if len(fields) < 2 {
            continue
        }
        num, _ := strconv.ParseFloat(fields[0], 64)
        unit := fields[1]
        switch {
        case strings.HasPrefix(unit, "week"):
            totalSeconds += num * 7 * 24 * 60 * 60
        case strings.HasPrefix(unit, "day"):
            totalSeconds += num * 24 * 60 * 60
        case strings.HasPrefix(unit, "hour"):
            totalSeconds += num * 60 * 60
        case strings.HasPrefix(unit, "minute"):
            totalSeconds += num * 60
        }
    }
    return totalSeconds
}
