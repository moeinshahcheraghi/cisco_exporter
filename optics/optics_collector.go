package optics

import (
	"log"
	"regexp"

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
    var iflistcmd, transceiverCmd string
    switch client.OSType {
    case rpc.IOS:
        iflistcmd = "show interfaces stats | exclude disabled"
        transceiverCmd = "show interfaces transceiver"
    case rpc.NXOS:
        iflistcmd = "show interface status | exclude disabled | exclude notconn | exclude sfpAbsent | exclude --------------------------------------------------------------------------------"
        transceiverCmd = "show interface transceiver details"
    case rpc.IOSXE:
        iflistcmd = "show interfaces stats | exclude disabled"
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

    if client.OSType == rpc.IOS || client.OSType == rpc.NXOS {
        out, err = client.RunCommand(transceiverCmd)
        if err != nil {
            if client.Debug {
                log.Printf("Transceiver command on %s: %s\n", labelValues[0], err.Error())
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
        for _, i := range interfaces {
            if optic, ok := opticsData[i]; ok {
                l := append(labelValues, i)
                ch <- prometheus.MustNewConstMetric(opticsTXDesc, prometheus.GaugeValue, optic.TxPower, l...)
                ch <- prometheus.MustNewConstMetric(opticsRXDesc, prometheus.GaugeValue, optic.RxPower, l...)
            }
        }
    } else if client.OSType == rpc.IOSXE {
        xeDev := regexp.MustCompile(`\S(\d+)/(\d+)/(\d+)`)
        for _, i := range interfaces {
            matches := xeDev.FindStringSubmatch(i)
            if matches == nil {
                continue
            }
            out, err := client.RunCommand("show hw-module subslot " + matches[1] + "/" + matches[2] + " transceiver " + matches[3] + " status")
            if err != nil {
                if client.Debug {
                    log.Printf("Transceiver command on %s: %s\n", labelValues[0], err.Error())
                }
                continue
            }
            optic, err := c.ParseTransceiver(client.OSType, out)
            if err != nil {
                if client.Debug {
                    log.Printf("Transceiver data for %s: %s\n", labelValues[0], err.Error())
                }
                continue
            }
            l := append(labelValues, i)
            ch <- prometheus.MustNewConstMetric(opticsTXDesc, prometheus.GaugeValue, optic.TxPower, l...)
            ch <- prometheus.MustNewConstMetric(opticsRXDesc, prometheus.GaugeValue, optic.RxPower, l...)
        }
    }
    return nil
}