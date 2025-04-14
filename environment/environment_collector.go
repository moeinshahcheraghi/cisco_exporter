package environment

import (
	"log"

	"github.com/lwlcom/cisco_exporter/rpc"

	"github.com/lwlcom/cisco_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_environment_"

var (
	temperaturesDesc *prometheus.Desc
	powerSupplyDesc  *prometheus.Desc
)

func init() {
	l := []string{"target", "item"}
	temperaturesDesc = prometheus.NewDesc(prefix+"sensor_temp", "Sensor temperatures", l, nil)
	l = append(l, "status")
	powerSupplyDesc = prometheus.NewDesc(prefix+"power_up", "Status of power supplies (1 OK, 0 Something is wrong)", l, nil)
}

type environmentCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &environmentCollector{}
}

// Name returns the name of the collector
func (*environmentCollector) Name() string {
	return "Environment"
}

// Describe describes the metrics
func (*environmentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- temperaturesDesc
	ch <- powerSupplyDesc
}

func (c *environmentCollector) Parse(ostype string, output string) ([]EnvironmentItem, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return nil, errors.New("'show environment all' is not implemented for " + ostype)
	}

	items := []EnvironmentItem{}

	lines := strings.Split(output, "\n")
	var currentSwitch string
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Identify current switch for labeling (e.g., "Switch 1:")
		if strings.HasPrefix(line, "Switch ") && strings.Contains(line, ":") {
			currentSwitch = strings.Split(line, ":")[0]
			continue
		}

		// Inlet or Hotspot temperature
		if strings.HasPrefix(line, "Inlet Temperature Value:") || strings.HasPrefix(line, "Hotspot Temperature Value:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				tempStr := strings.TrimSpace(strings.Replace(parts[1], "Degree Celsius", "", 1))
				temp := util.Str2float64(tempStr)

				items = append(items, EnvironmentItem{
					Name:        currentSwitch + " " + name,
					IsTemp:      true,
					Temperature: temp,
				})
			}
		}

		// FAN or PSU status
		if strings.Contains(line, "FAN") && strings.Contains(line, "is OK") {
			items = append(items, EnvironmentItem{
				Name:   line,
				IsTemp: false,
				OK:     true,
				Status: "OK",
			})
		}

		// Parse PSU table (last section)
		if strings.HasPrefix(line, "1A") || strings.HasPrefix(line, "1B") || strings.HasPrefix(line, "2A") || strings.HasPrefix(line, "2B") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				name := fields[0] + " " + fields[1]
				status := fields[3]
				ok := status == "OK"
				items = append(items, EnvironmentItem{
					Name:   name,
					IsTemp: false,
					OK:     ok,
					Status: status,
				})
			}
		}
	}
	return items, nil
}
