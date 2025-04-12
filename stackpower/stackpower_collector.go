package stackpower

import (
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const prefix = "cisco_stackpower_"

var (
	totalPowerDesc  = prometheus.NewDesc(prefix+"total_power_watts", "Total stack power in watts", []string{"target", "stack_name", "mode", "topology"}, nil)
	rsvdPowerDesc   = prometheus.NewDesc(prefix+"reserved_power_watts", "Reserved stack power in watts", []string{"target", "stack_name"}, nil)
	allocPowerDesc  = prometheus.NewDesc(prefix+"allocated_power_watts", "Allocated stack power in watts", []string{"target", "stack_name"}, nil)
	unusedPowerDesc = prometheus.NewDesc(prefix+"unused_power_watts", "Unused stack power in watts", []string{"target", "stack_name"}, nil)
	switchCountDesc = prometheus.NewDesc(prefix+"switch_count", "Number of switches in stack", []string{"target", "stack_name"}, nil)
	psCountDesc     = prometheus.NewDesc(prefix+"power_supply_count", "Number of power supplies in stack", []string{"target", "stack_name"}, nil)
)

type stackPowerCollector struct{}

func NewCollector() collector.RPCCollector {
	return &stackPowerCollector{}
}

func (*stackPowerCollector) Name() string { return "StackPower" }

func (*stackPowerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- totalPowerDesc
	ch <- rsvdPowerDesc
	ch <- allocPowerDesc
	ch <- unusedPowerDesc
	ch <- switchCountDesc
	ch <- psCountDesc
}

func (*stackPowerCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labels []string) error {
	out, err := client.RunCommand("show stack-power")
	if err != nil {
		log.Errorln("StackPower command failed:", err)
		return err
	}

	stacks, err := ParseStackPower(out)
	if err != nil {
		log.Errorln("StackPower parsing error:", err)
		return err
	}

	for _, stack := range stacks {
		l := []string{labels[0], stack.Name}

		ch <- prometheus.MustNewConstMetric(totalPowerDesc, prometheus.GaugeValue, stack.TotalPower, labels[0], stack.Name, stack.Mode, stack.Topology)
		ch <- prometheus.MustNewConstMetric(rsvdPowerDesc, prometheus.GaugeValue, stack.ReservedPower, l...)
		ch <- prometheus.MustNewConstMetric(allocPowerDesc, prometheus.GaugeValue, stack.AllocatedPower, l...)
		ch <- prometheus.MustNewConstMetric(unusedPowerDesc, prometheus.GaugeValue, stack.UnusedPower, l...)
		ch <- prometheus.MustNewConstMetric(switchCountDesc, prometheus.GaugeValue, float64(stack.NumSwitches), l...)
		ch <- prometheus.MustNewConstMetric(psCountDesc, prometheus.GaugeValue, float64(stack.NumPowerSupplies), l...)
	}

	return nil
}
