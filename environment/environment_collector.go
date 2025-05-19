package environment

import (
    "fmt"
    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_environment_"

var (
    statusDesc         *prometheus.Desc
    temperatureDesc    *prometheus.Desc
    slotTemperatureDesc *prometheus.Desc
)

func init() {
    l := []string{"target", "name"}
    statusDesc = prometheus.NewDesc(prefix+"status", "Status of environment item", l, nil)
    temperatureDesc = prometheus.NewDesc(prefix+"temperature_celsius", "Temperature of environment item", l, nil)
    slotTemperatureDesc = prometheus.NewDesc(prefix+"slot_temperature_celsius", "Temperature of hardware slot", append(l, "slot"), nil)
}

type environmentCollector struct{}

func NewCollector() collector.RPCCollector {
    return &environmentCollector{}
}

func (*environmentCollector) Name() string {
    return "Environment"
}

func (*environmentCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- statusDesc
    ch <- temperatureDesc
    ch <- slotTemperatureDesc
}

func (c *environmentCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
    out, err := client.RunCommand("show environment")
    if err != nil {
        return err
    }
    items, err := Parse(client.OSType, out)
    if err != nil {
        return err
    }
    for _, item := range items {
        l := append(labelValues, item.Name)
        if item.IsTemp {
            ch <- prometheus.MustNewConstMetric(temperatureDesc, prometheus.GaugeValue, item.Temperature, l...)
        } else {
            val := 0.0
            if item.OK {
                val = 1.0
            }
            ch <- prometheus.MustNewConstMetric(statusDesc, prometheus.GaugeValue, val, l...)
        }
    }

    if client.OSType == rpc.IOSXE {
        for slot := 0; slot < 10; slot++ {
            cmd := fmt.Sprintf("show platform hardware slot %d env temperature", slot)
            out, err := client.RunCommand(cmd)
            if err != nil {
                continue
            }
            item, err := ParseSlotTemperature(client.OSType, out, fmt.Sprintf("%d", slot))
            if err != nil {
                continue
            }
            l := append(labelValues, item.Name, item.Slot)
            ch <- prometheus.MustNewConstMetric(slotTemperatureDesc, prometheus.GaugeValue, item.Temperature, l...)
        }
    }
    return nil
}