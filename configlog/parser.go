package configlog

import (
	"errors"
	"strings"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
)

func (c *configlogCollector) Parse(ostype string, output string) (int, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return 0, errors.New("'show logging | include Config' not implemented for " + ostype)
	}
	lines := strings.Split(output, "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count, nil
}