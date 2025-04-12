package environment

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	tempSwitchRe    = regexp.MustCompile(`Switch (\d+): SYSTEM TEMPERATURE is (\w+)`)
	tempValueRe     = regexp.MustCompile(`(Inlet|Hotspot) Temperature Value:\s+(\d+)`)
	tempYellowRe    = regexp.MustCompile(`Yellow Threshold : (\d+)`)
	tempRedRe       = regexp.MustCompile(`Red Threshold\s+: (\d+)`)
	tempStateRe     = regexp.MustCompile(`Temperature State:\s+(\w+)`)
)

func ParseTemperature(device string, output string) ([]prometheus.Metric, error) {
	var metrics []prometheus.Metric
	var currentSwitch string
	var currentType string

	scanner := bufio.NewScanner(strings.NewReader(output))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if switchMatch := tempSwitchRe.FindStringSubmatch(line); switchMatch != nil {
			currentSwitch = switchMatch[1]
			continue
		}

		if typeMatch := tempValueRe.FindStringSubmatch(line); typeMatch != nil {
			currentType = strings.ToLower(typeMatch[1])
			value, _ := strconv.ParseFloat(typeMatch[2], 64)
			
			// Create metrics
			metrics = append(metrics, prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"cisco_environment_temperature_value",
					"Current temperature value",
					[]string{"device", "switch", "type"},
					nil,
				),
				prometheus.GaugeValue,
				value,
				device, currentSwitch, currentType,
			))
		}

		if yellowMatch := tempYellowRe.FindStringSubmatch(line); yellowMatch != nil {
			value, _ := strconv.ParseFloat(yellowMatch[1], 64)
			metrics = append(metrics, prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"cisco_environment_temperature_threshold_yellow",
					"Yellow threshold temperature",
					[]string{"device", "switch", "type"},
					nil,
				),
				prometheus.GaugeValue,
				value,
				device, currentSwitch, currentType,
			))
		}

		if redMatch := tempRedRe.FindStringSubmatch(line); redMatch != nil {
			value, _ := strconv.ParseFloat(redMatch[1], 64)
			metrics = append(metrics, prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"cisco_environment_temperature_threshold_red",
					"Red threshold temperature",
					[]string{"device", "switch", "type"},
					nil,
				),
				prometheus.GaugeValue,
				value,
				device, currentSwitch, currentType,
			))
		}

		if stateMatch := tempStateRe.FindStringSubmatch(line); stateMatch != nil {
			state := 0.0
			if stateMatch[1] == "GREEN" {
				state = 1.0
			}
			metrics = append(metrics, prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"cisco_environment_temperature_state",
					"Temperature state (1 = GREEN)",
					[]string{"device", "switch", "type"},
					nil,
				),
				prometheus.GaugeValue,
				state,
				device, currentSwitch, currentType,
			))
		}
	}
	
	return metrics, nil
}
