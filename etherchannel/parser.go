package etherchannel

import (
    "errors"
    "regexp"
    "strconv"
    "strings"

    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
)

func Parse(ostype string, output string) (int, []EtherChannelGroup, error) {
    if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
        return 0, nil, errors.New("'show etherchannel summary' is not implemented for " + ostype)
    }

    groupsTotalRegexp := regexp.MustCompile(`Number of channel-groups in use:\s+(\d+)`)
    groupRegexp := regexp.MustCompile(`^(\d+)\s+(\S+)\((\S+)\)\s+(\S+)\s+(.*)$`)
    portRegexp := regexp.MustCompile(`(\S+)\((\S+)\)`)

    lines := strings.Split(output, "\n")
    var groupsTotal int
    groups := []EtherChannelGroup{}
    for _, line := range lines {
        if matches := groupsTotalRegexp.FindStringSubmatch(line); matches != nil {
            var err error
            groupsTotal, err = strconv.Atoi(matches[1])
            if err != nil {
                return 0, nil, err
            }
        } else if matches := groupRegexp.FindStringSubmatch(line); matches != nil {
            group := EtherChannelGroup{
                Group:       matches[1],
                PortChannel: matches[2],
                Status:      matches[3],
                Protocol:    matches[4],
            }
            portsStr := matches[5]
            portMatches := portRegexp.FindAllStringSubmatch(portsStr, -1)
            for _, pm := range portMatches {
                group.Ports = append(group.Ports, EtherChannelPort{
                    Port:   pm[1],
                    Status: pm[2],
                })
            }
            groups = append(groups, group)
        }
    }
    return groupsTotal, groups, nil
}
