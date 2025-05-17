package cef

import (
	"errors"
	"regexp"
	"strings"
	"strconv"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
)

func (c *cefCollector) ParseInterfaces(ostype string, output string) ([]string, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return nil, errors.New("'show interfaces stats' not implemented for " + ostype)
	}
	var items []string
	re := regexp.MustCompile(`^([a-zA-Z0-9\/\.-]+)\s*`)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		items = append(items, matches[1])
	}
	return items, nil
}

func (c *cefCollector) Parse(ostype string, output string) (int, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return 0, errors.New("'show cef interface' not implemented for " + ostype)
	}
	re := regexp.MustCompile(`(\d+) packets dropped`)
	matches := re.FindStringSubmatch(output)
	if matches == nil {
		return 0, nil
	}
	return strconv.Atoi(matches[1])
}