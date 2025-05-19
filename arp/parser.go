package arp

import (
    "errors"
    "regexp"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/moeinshahcheraghi/cisco_exporter/util"
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
    return util.Str2int(matches[1]), nil
}