package loginfail

import (
	"errors"
	"regexp"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
)

func (c *loginfailCollector) Parse(ostype string, output string) (int, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return 0, errors.New("'show login failures' not implemented for " + ostype)
	}
	re := regexp.MustCompile(`Total Failed Attempts: (\d+)`)
	matches := re.FindStringSubmatch(output)
	if matches == nil {
		return 0, nil // No failures found
	}
	count, _ := strconv.Atoi(matches[1])
	return count, nil
}