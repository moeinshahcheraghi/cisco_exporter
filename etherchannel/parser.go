package etherchannel

import (
	"errors"
	"regexp"
	"strings"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
)

type EtherChannelItem struct {
	Group     string
	Protocol  string
	Up        bool
	PortsCount int
}

func (c *etherchannelCollector) Parse(ostype string, output string) ([]EtherChannelItem, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return nil, errors.New("'show etherchannel summary' not implemented for " + ostype)
	}
	items := []EtherChannelItem{}
	re := regexp.MustCompile(`(\d+)\s+(\S+)\s+(\S+)\s+(\S+)`)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		ports := strings.Count(matches[4], "P") + strings.Count(matches[4], "D") // Count active ports
		items = append(items, EtherChannelItem{
			Group:     matches[1],
			Protocol:  matches[3],
			Up:        strings.Contains(matches[4], "P"),
			PortsCount: ports,
		})
	}
	return items, nil
}