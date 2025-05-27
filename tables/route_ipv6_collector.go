package tables

import (
	"regexp"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/moeinshahcheraghi/cisco_exporter/util"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	routesIPv6Desc = prometheus.NewDesc(prefix+"routes_ipv6", "Number of IPv6 routes", []string{"target"}, nil)
)

type routeIPv6Collector struct{}

func NewRouteIPv6Collector() collector.RPCCollector {
	return &routeIPv6Collector{}
}

func (*routeIPv6Collector) Name() string {
	return "TablesRouteIPv6"
}

func (*routeIPv6Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- routesIPv6Desc
}

func (c *routeIPv6Collector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show ipv6 route summary")
	if err != nil {
		return err
	}
	routeCount := parseRoutesIPv6(out)
	ch <- prometheus.MustNewConstMetric(routesIPv6Desc, prometheus.GaugeValue, routeCount, labelValues...)
	return nil
}

func parseRoutesIPv6(output string) float64 {
	re := regexp.MustCompile(`Total\s+(\d+)`)
	matches := re.FindStringSubmatch(output)
	if matches != nil {
		return util.Str2float64(matches[1])
	}
	return 0
}