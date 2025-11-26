package optics

import (
	"log"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_optics_"

var (
	opticsTXDesc *prometheus.Desc
	opticsRXDesc *prometheus.Desc
)

func init() {
	l := []string{"target", "interface"}
	opticsTXDesc = prometheus.NewDesc(prefix+"tx", "Transceiver Tx power (dBm)", l, nil)
	opticsRXDesc = prometheus.NewDesc(prefix+"rx", "Transceiver Rx power (dBm)", l, nil)
}

type opticsCollector struct{}

func NewCollector() collector.RPCCollector {
	return &opticsCollector{}
}

func (*opticsCollector) Name() string {
	return "Optics"
}

func (*opticsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- opticsTXDesc
	ch <- opticsRXDesc
}

func (c *opticsCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	// دستور یکسان برای گرفتن لیست اینترفیس‌ها
	iflistcmd := "show interfaces status | exclude disabled"
	if client.OSType == rpc.IOSXE || client.OSType == rpc.IOS {
		iflistcmd = "show interfaces status | exclude disabled"
	}

	out, err := client.RunCommand(iflistcmd)
	if err != nil {
		return err
	}

	interfaces, err := c.ParseInterfaces(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("ParseInterfaces for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}

	// === بخش مهم: استفاده از دستور bulk برای همه OS ها ===
	var transceiverCmd string
	switch client.OSType {
	case rpc.NXOS:
		transceiverCmd = "show interface transceiver details"
	default: // IOS + IOS-XE
		transceiverCmd = "show interfaces transceiver"
		// اگر detail بهتر کار کرد، این را بگذارید:
		// transceiverCmd = "show interfaces transceiver detail"
	}

	out, err = client.RunCommand(transceiverCmd)
	if err != nil {
		if client.Debug {
			log.Printf("Bulk transceiver command failed on %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}

	opticsData, err := c.ParseAllTransceivers(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("ParseAllTransceivers for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}

	// فقط اینترفیس‌هایی که ترانسیور دارند را گزارش کن
	for _, iface := range interfaces {
		if optic, exists := opticsData[iface]; exists {
			l := append(labelValues, iface)
			ch <- prometheus.MustNewConstMetric(opticsTXDesc, prometheus.GaugeValue, optic.TxPower, l...)
			ch <- prometheus.MustNewConstMetric(opticsRXDesc, prometheus.GaugeValue, optic.RxPower, l...)
		}
	}

	return nil
}