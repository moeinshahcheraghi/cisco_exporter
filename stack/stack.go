package stack

import "github.com/prometheus/client_golang/prometheus"

var (
	stackPortStatus *prometheus.GaugeVec
)

// Describe sends the super-set of all possible descriptors of metrics
func Describe(ch chan<- *prometheus.Desc) {
	stackPortStatus.Describe(ch)
}

// Collect is called by the Prometheus registry when collecting metrics
func Collect(ch chan<- prometheus.Metric) {
	stackPortStatus.Collect(ch)
}

func init() {
	stackPortStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cisco_stack_port_status",
			Help: "Status of stack ports (1 = OK, 0 = Bad)",
		},
		[]string{"device", "switch", "port"},
	)
}
