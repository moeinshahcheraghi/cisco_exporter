package slottemp

import (
	"errors"
	"regexp"
	"strings"
	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/moeinshahcheraghi/cisco_exporter/util"
)

type SlotTempItem struct {
	Slot        string
	Temperature float64
}

func (c *slottempCollector) Parse(ostype string, output string) ([]SlotTempItem, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return nil, errors.New("'show environment all' not implemented for " + ostype)
	}
	items := []SlotTempItem{}
	re := regexp.MustCompile(`Slot (\d+) Temperature:\s+(\d+\.\d+) C`)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		items = append(items, SlotTempItem{
			Slot:        matches[1],
			Temperature: util.Str2float64(matches[2]),
		})
	}
	return items, nil
}