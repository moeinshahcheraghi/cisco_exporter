package optics

import (
	"log"
	"regexp"
	"strings"

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
	opticsTXDesc = prometheus.NewDesc(prefix+"tx", "Transceiver Tx power", l, nil)
	opticsRXDesc = prometheus.NewDesc(prefix+"rx", "Transceiver Rx power", l, nil)
}

type opticsCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &opticsCollector{}
}

// Name returns the name of the collector
func (*opticsCollector) Name() string {
	return "Optics"
}

// Describe describes the metrics
func (*opticsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- opticsTXDesc
	ch <- opticsRXDesc
}

// Collect collects metrics from Cisco
func (c *opticsCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	var cmd string
	switch client.OSType {
	case rpc.IOS, rpc.IOSXE:
		cmd = "show interfaces transceiver"
	case rpc.NXOS:
		cmd = "show interface transceiver details"
	}
	out, err := client.RunCommand(cmd)
	if err != nil {
		return err
	}
	items, err := c.Parse(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("Parse optics for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}
	for _, item := range items {
		l := append(labelValues, item.Interface)
		ch <- prometheus.MustNewConstMetric(opticsTXDesc, prometheus.GaugeValue, float64(item.TxPower), l...)
		ch <- prometheus.MustNewConstMetric(opticsRXDesc, prometheus.GaugeValue, float64(item.RxPower), l...)
	}
	return nil
}

// Parse parses cli output and tries to find tx/rx power for all interfaces
func (c *opticsCollector) Parse(ostype string, output string) ([]Optics, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return nil, errors.New("Transceiver data is not implemented for " + ostype)
	}
	items := []Optics{}
	lines := strings.Split(output, "\n")
	re := regexp.MustCompile(`(\S+)\s+((?:-)?\d+\.\d+)\s+((?:-)?\d+\.\d+)`)
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		items = append(items, Optics{
			Interface: matches[1],
			TxPower:   util.Str2float64(matches[2]),
			RxPower:   util.Str2float64(matches[3]),
		})
	}
	return items, nil
}