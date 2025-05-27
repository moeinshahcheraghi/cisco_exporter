package tables

import (
	"regexp"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/moeinshahcheraghi/cisco_exporter/util"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	routesIPv4Desc = prometheus.NewDesc(prefix+"routes_ipv4", "Number of IPv4 routes", []string{"target"}, nil)
)

type routeIPv4Collector struct{}

func NewRouteIPv4Collector() collector.RPCCollector {
	return &routeIPv4Collector{}
}

func (*routeIPv4Collector) Name() string {
	return "TablesRouteIPv4"
}

func (*routeIPv4Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- routesIPv4Desc
}

func (c *routeIPv4Collector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show ip route summary")
	if err != nil {
		return err
	}
	routeCount := parseRoutesIPv4(out)
	ch <- prometheus.MustNewConstMetric(routesIPv4Desc, prometheus.GaugeValue, routeCount, labelValues...)
	return nil
}

func parseRoutesIPv4(output string) float64 {
	re := regexp.MustCompile(`Total\s+(\d+)\s+(\d+)`)
	matches := re.FindStringSubmatch(output)
	if matches != nil {
		networks := util.Str2float64(matches[1])
		subnets := util.Str2float64(matches[2])
		return networks + subnets
	}
	return 0
}