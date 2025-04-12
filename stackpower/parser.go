package stackpower

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)


func ParseStackPower(output string) ([]StackPower, error) {
	lines := strings.Split(output, "\n")

	if len(lines) < 2 {
		return nil, errors.New("unexpected stack power output format")
	}

	var stacks []StackPower
	// Skip headers and separator
	re := regexp.MustCompile(`^([\w\-]+)\s+(\w+)\s+(\w+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		totalPower, _ := strconv.ParseFloat(matches[4], 64)
		rsvdPower, _ := strconv.ParseFloat(matches[5], 64)
		allocPower, _ := strconv.ParseFloat(matches[6], 64)
		unusedPower, _ := strconv.ParseFloat(matches[7], 64)
		numSwitches, _ := strconv.Atoi(matches[8])
		numPS, _ := strconv.Atoi(matches[9])

		stack := StackPower{
			Name:             matches[1],
			Mode:             matches[2],
			Topology:         matches[3],
			TotalPower:       totalPower,
			ReservedPower:    rsvdPower,
			AllocatedPower:   allocPower,
			UnusedPower:      unusedPower,
			NumSwitches:      numSwitches,
			NumPowerSupplies: numPS,
		}

		stacks = append(stacks, stack)
	}

	if len(stacks) == 0 {
		return nil, errors.New("no stack power entries found")
	}

	return stacks, nil
}
