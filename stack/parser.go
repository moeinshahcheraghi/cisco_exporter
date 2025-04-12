package stack

import (
	"bufio"
	"regexp"
	"strings"
)

var (
	stackPortHeaderRe = regexp.MustCompile(`Switch#\s+Port1\s+Port2`)
	stackPortLineRe   = regexp.MustCompile(`^\d+\s+(\S+)\s+(\S+)\s*$`)
)

func ParseStackPorts(device string, output string) ([]prometheus.Metric, error) {
	var metrics []prometheus.Metric

	scanner := bufio.NewScanner(strings.NewReader(output))
	headerFound := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if stackPortHeaderRe.MatchString(line) {
			headerFound = true
			continue
		}
		
		if !headerFound || line == "" {
			continue
		}

		matches := stackPortLineRe.FindStringSubmatch(line)
		if matches != nil {
			switchNum := strings.Split(line, " ")[0]
			port1Status := 0.0
			if matches[1] == "OK" {
				port1Status = 1.0
			}
			port2Status := 0.0
			if matches[2] == "OK" {
				port2Status = 1.0
			}

			metrics = append(metrics, stackPortStatus.WithLabelValues(
				device, switchNum, "1").Set(port1Status)
			metrics = append(metrics, stackPortStatus.WithLabelValues(
				device, switchNum, "2").Set(port2Status)
		}
	}

	return metrics, nil
}
