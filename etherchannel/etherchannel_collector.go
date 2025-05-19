package etherchannel

import (
    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_etherchannel_"

var (
    groupsTotalDesc      *prometheus.Desc
    groupUpDesc          *prometheus.Desc
    groupActivePortsDesc *prometheus.Desc
)

func init() {
    l := []string{"target"}
    groupsTotalDesc = prometheus.NewDesc(prefix+"groups_total", "Total number of channel-groups in use", l, nil)
    groupUpDesc = prometheus.NewDesc(prefix+"group_up", "Whether the group is in use (1 if U, 0 otherwise)", append(l, "group", "port_channel", "protocol"), nil)
    groupActivePortsDesc = prometheus.NewDesc(prefix+"group_active_ports", "Number of active ports in the group", append(l, "group", "port_channel"), nil)
}

type etherchannelCollector struct{}

func NewCollector() collector.RPCCollector {
    return &etherchannelCollector{}
}

func (*etherchannelCollector) Name() string {
    return "Etherchannel"
}

func (*etherchannelCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- groupsTotalDesc
    ch <- groupUpDesc
    ch <- groupActivePortsDesc
}

func (c *etherchannelCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
    out, err := client.RunCommand("show etherchannel summary")
    if err != nil {
        return err
    }
    groupsTotal, groups, err := Parse(client.OSType, out)
    if err != nil {
        return err
    }
    ch <- prometheus.MustNewConstMetric(groupsTotalDesc, prometheus.GaugeValue, float64(groupsTotal), labelValues...)

    for _, group := range groups {
        l := append(labelValues, group.Group, group.PortChannel, group.Protocol)
        up := 0.0
        if group.Status == "U" {
            up = 1.0
        }
        ch <- prometheus.MustNewConstMetric(groupUpDesc, prometheus.GaugeValue, up, l...)

        activePorts := 0
        for _, port := range group.Ports {
            if port.Status == "P" {
                activePorts++
            }
        }
        ch <- prometheus.MustNewConstMetric(groupActivePortsDesc, prometheus.GaugeValue, float64(activePorts), append(labelValues, group.Group, group.PortChannel)...)
    }
    return nil
}