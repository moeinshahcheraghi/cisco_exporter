package stackport

import (
	"errors"
	"strings"

	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
)

type StackPortItem struct {
	Switch string
	Port   string
	OK     bool
}

func Parse(ostype string, output string) ([]StackPortItem, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return nil, errors.New("'show switch stack-ports' is not implemented for " + ostype)
	}

	items := []StackPortItem{}
	lines := strings.Split(output, "\n")
	separatorFound := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "----------------------------") {
			separatorFound = true
			continue
		}
		if !separatorFound {
			continue
		}
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		switchNum := fields[0]
		port1Status := fields[1]
		port2Status := fields[2]

		items = append(items, StackPortItem{
			Switch: switchNum,
			Port:   "Port1",
			OK:     port1Status == "OK",
		}, StackPortItem{
			Switch: switchNum,
			Port:   "Port2",
			OK:     port2Status == "OK",
		})
	}

	return items, nil
}