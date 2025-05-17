package slottemp

import (
	"log"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_slot_temperature_"

var (
	tempDesc *prometheus.Desc
)

func init() {
	l := []string{"target", "slot"}
	tempDesc = prometheus.NewDesc(prefix+"celsius", "Temperature in Celsius", l, nil)
}

type slottempCollector struct{}

func NewCollector() collector.RPCCollector {
	return &slottempCollector{}
}

func (*slottempCollector) Name() string {
	return "SlotTemperature"
}

func (*slottempCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- tempDesc
}

func (c *slottempCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show environment all") // Using a summary command
	if err != nil {
		return err
	}
	items, err := c.Parse(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("Parse slot temp for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}
	for _, item := range items {
		l := append(labelValues, item.Slot)
		ch <- prometheus.MustNewConstMetric(tempDesc, prometheus.GaugeValue, item.Temperature, l...)
	}
	return nil
}