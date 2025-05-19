package arp

import (
    "errors"
    "regexp"
    "strconv"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
)

func Parse(ostype string, output string) (int, error) {
    if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
        return 0, errors.New("'show arp summary' is not implemented for " + ostype)
    }
    totalRegexp := regexp.MustCompile(`Total\s+ARP\s+entries:\s+(\d+)`)
    matches := totalRegexp.FindStringSubmatch(output)
    if matches == nil {
        return 0, errors.New("Total ARP entries not found")
    }
    total, err := strconv.Atoi(matches[1])
    if err != nil {
        return 0, err
    }
    return total, nil
}