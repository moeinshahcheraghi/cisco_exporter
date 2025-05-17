package arp

import (
	"errors"
	"regexp"
	"strconv"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
)

func (c *arpCollector) Parse(ostype string, output string) (int, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return 0, errors.New("'show arp summary' not implemented for " + ostype)
	}
	re := regexp.MustCompile(`Total number of entries:\s*(\d+)`)
	matches := re.FindStringSubmatch(output)
	if matches == nil {
		return 0, errors.New("Could not find ARP entries")
	}
	return strconv.Atoi(matches[1])
}