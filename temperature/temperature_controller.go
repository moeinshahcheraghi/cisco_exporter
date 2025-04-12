package temperature

import (
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"regexp"
	"strconv"
	"strings"
)

var (
	inletTempDesc = prometheus.NewDesc("cisco_temperature_inlet_celsius", "Inlet temperature in Celsius", []string{"target", "switch"}, nil)
	hotspotTempDesc = prometheus.NewDesc("cisco_temperature_hotspot_celsius", "Hotspot temperature in Celsius", []string{"target", "switch"}, nil)
)

type temperatureCollector struct{}

func NewCollector() collector.RPCCollector {
	return &temperatureCollector{}
}

func (c *temperatureCollector) Name() string {
	return "Temperature"
}

func (c *temperatureCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- inletTempDesc
	ch <- hotspotTempDesc
}

func (c *temperatureCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labels []string) error {
	out, err := client.RunCommand("show environment temperature")
	if err != nil {
		return err
	}

	switchRe := regexp.MustCompile(`Switch (\d+):`)
	inletRe := regexp.MustCompile(`Inlet Temperature Value: (\d+)`)
	hotspotRe := regexp.MustCompile(`Hotspot Temperature Value: (\d+)`)

	lines := strings.Split(out, "\n")
	var currentSwitch string
	for _, line := range lines {
		if m := switchRe.FindStringSubmatch(line); m != nil {
			currentSwitch = m[1]
		} else if m := inletRe.FindStringSubmatch(line); m != nil {
			val, _ := strconv.ParseFloat(m[1], 64)
			ch <- prometheus.MustNewConstMetric(inletTempDesc, prometheus.GaugeValue, val, labels[0], currentSwitch)
		} else if m := hotspotRe.FindStringSubmatch(line); m != nil {
			val, _ := strconv.ParseFloat(m[1], 64)
			ch <- prometheus.MustNewConstMetric(hotspotTempDesc, prometheus.GaugeValue, val, labels[0], currentSwitch)
		}
	}

	return nil
}
