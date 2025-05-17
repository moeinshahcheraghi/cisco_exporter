package etherchannel

import (
	"log"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_etherchannel_"

var (
	channelsTotalDesc *prometheus.Desc
	statusDesc        *prometheus.Desc
	portsCountDesc    *prometheus.Desc
)

func init() {
	l := []string{"target", "group", "protocol"}
	channelsTotalDesc = prometheus.NewDesc(prefix+"total", "Total number of EtherChannels", []string{"target"}, nil)
	statusDesc = prometheus.NewDesc(prefix+"status", "EtherChannel status (1 = up, 0 = down)", l, nil)
	portsCountDesc = prometheus.NewDesc(prefix+"ports_count", "Number of ports in EtherChannel", l, nil)
}

type etherchannelCollector struct{}

func NewCollector() collector.RPCCollector {
	return &etherchannelCollector{}
}

func (*etherchannelCollector) Name() string {
	return "EtherChannel"
}

func (*etherchannelCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- channelsTotalDesc
	ch <- statusDesc
	ch <- portsCountDesc
}

func (c *etherchannelCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show etherchannel summary")
	if err != nil {
		return err
	}
	items, err := c.Parse(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("Parse EtherChannel for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}
	ch <- prometheus.MustNewConstMetric(channelsTotalDesc, prometheus.GaugeValue, float64(len(items)), labelValues[0])
	for _, item := range items {
		l := append(labelValues, item.Group, item.Protocol)
		status := 0.0
		if item.Up {
			status = 1.0
		}
		ch <- prometheus.MustNewConstMetric(statusDesc, prometheus.GaugeValue, status, l...)
		ch <- prometheus.MustNewConstMetric(portsCountDesc, prometheus.GaugeValue, float64(item.PortsCount), l...)
	}
	return nil
}