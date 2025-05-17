package spanningtree

import (
	"errors"
	"regexp"
	"strings"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
)

type SpanningTreeItem struct {
	VLAN      string
	Interface string
	State     int // 1 = Forwarding, 0 = Blocking, etc.
}

func (c *spanningtreeCollector) Parse(ostype string, output string) ([]SpanningTreeItem, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return nil, errors.New("'show spanning-tree detail' not implemented for " + ostype)
	}
	items := []SpanningTreeItem{}
	re := regexp.MustCompile(`VLAN(\d+).*?(\S+) is (\S+)`)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		state := 0
		if matches[3] == "Forwarding" {
			state = 1
		}
		items = append(items, SpanningTreeItem{
			VLAN:      matches[1],
			Interface: matches[2],
			State:     state,
		})
	}
	return items, nil
}