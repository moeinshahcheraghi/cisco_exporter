package environment

import (
	"errors"
	"regexp"
	"strings"

	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/moeinshahcheraghi/cisco_exporter/util"
)

// Parse dispatches to the OS-specific parser
func Parse(ostype string, output string) ([]EnvironmentItem, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return nil, errors.New("'show environment' is not implemented for " + ostype)
	}

	if ostype == rpc.NXOS {
		return parseNXOS(output)
	}

	return parseIOS(output)
}

// parseIOS parses 'show environment all' output from IOS / IOS-XE devices
func parseIOS(output string) ([]EnvironmentItem, error) {
	items := []EnvironmentItem{}
	lines := strings.Split(output, "\n")
	var currentSwitch string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detect Switch N:
		if strings.HasPrefix(line, "Switch ") && strings.Contains(line, ":") {
			currentSwitch = strings.Split(line, ":")[0]
			continue
		}

		// Inlet / Hotspot temperatures
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

		// FAN or PSU status lines
		if strings.Contains(line, "FAN") && strings.Contains(line, "is OK") {
			items = append(items, EnvironmentItem{
				Name:   line,
				IsTemp: false,
				OK:     true,
				Status: "OK",
			})
		}

		// PSU table at the bottom
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

// parseNXOS parses NX-OS 'show environment' output. NX-OS prints three
// separate sections (Power Supply, Fan, Temperature), each with its own
// table. We track which section we're in and apply a matching regex per
// section rather than relying on fixed column positions, since column
// widths vary slightly between platforms (N9K/N7K/N5K/...).
//
// Example section headers/rows this expects:
//
//	Power Supply:
//	Voltage: 12 Volts
//	Power                                         Actual
//	Supply   Model                Output          Power     Status
//	Number                        (Watts)         Capacity
//	1        NXA-PAC-750W-PI      NA              750 W      Ok
//
//	Fan:
//	------------------------------------------------------
//	Fan               Model                Hw    Direction   Status
//	------------------------------------------------------
//	Fan1(sys_fan1)     NXA-FAN-30CFM-B      --    front-to-back Ok
//
//	Temperature:
//	--------------------------------------------------------------
//	Module   Sensor        MajorThresh   MinorThres   CurTemp     Status
//	--------------------------------------------------------------
//	1        Ambient        70            60             35        Ok
func parseNXOS(output string) ([]EnvironmentItem, error) {
	items := []EnvironmentItem{}
	lines := strings.Split(output, "\n")

	// last whitespace-separated token is always the status word (Ok, Failed,
	// Faulty, Absent, Shutdown, Powered-Off, ...)
	powerRe := regexp.MustCompile(`^(\d+)\s+(\S.*\S)\s{2,}\S.*\s(\S+)$`)
	fanRe := regexp.MustCompile(`^(\S.*\S)\s{2,}\S.*\s(\S+)$`)
	tempRe := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(-?\d+)\s+(-?\d+)\s+(-?\d+)\s+(\S+)$`)

	section := ""
	for _, raw := range lines {
		line := strings.TrimRight(raw, " \t\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "Power Supply"):
			section = "power"
			continue
		case trimmed == "Fan:" || strings.HasPrefix(trimmed, "Fan:"):
			section = "fan"
			continue
		case strings.HasPrefix(trimmed, "Temperature"):
			section = "temp"
			continue
		case strings.HasPrefix(trimmed, "----"),
			strings.HasPrefix(trimmed, "Voltage:"),
			strings.HasPrefix(trimmed, "Supply") && strings.Contains(trimmed, "Model"),
			strings.HasPrefix(trimmed, "Number"),
			strings.HasPrefix(trimmed, "Module") && strings.Contains(trimmed, "Sensor"),
			strings.HasPrefix(trimmed, "Fan") && strings.Contains(trimmed, "Model"):
			// table headers / separators, skip
			continue
		}

		switch section {
		case "power":
			if m := powerRe.FindStringSubmatch(line); m != nil {
				status := m[3]
				items = append(items, EnvironmentItem{
					Name:   "PowerSupply " + m[1] + " " + strings.TrimSpace(m[2]),
					IsTemp: false,
					OK:     strings.EqualFold(status, "Ok"),
					Status: status,
				})
			}
		case "fan":
			if m := fanRe.FindStringSubmatch(line); m != nil {
				status := m[2]
				items = append(items, EnvironmentItem{
					Name:   strings.TrimSpace(m[1]),
					IsTemp: false,
					OK:     strings.EqualFold(status, "Ok"),
					Status: status,
				})
			}
		case "temp":
			if m := tempRe.FindStringSubmatch(line); m != nil {
				temp := util.Str2float64(m[5])
				items = append(items, EnvironmentItem{
					Name:        "Module " + m[1] + " " + m[2],
					IsTemp:      true,
					Temperature: temp,
				})
			}
		}
	}

	return items, nil
}