package spanningtree

import (
    "errors"
    "regexp"
    "strings"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
)

func Parse(ostype string, output string) ([]SpanningTreeInstance, error) {
    if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
        return nil, errors.New("'show spanning-tree detail' is not implemented for " + ostype)
    }
    instanceRegexp := regexp.MustCompile(`Spanning tree instance\s+(\S+)`)
    blockedRegexp := regexp.MustCompile(`port\s+\S+\s+blocking`)
    lines := strings.Split(output, "\n")
    instances := make(map[string]int)
    currentInstance := ""
    for _, line := range lines {
        if matches := instanceRegexp.FindStringSubmatch(line); matches != nil {
            currentInstance = matches[1]
        }
        if blockedRegexp.MatchString(line) && currentInstance != "" {
            instances[currentInstance]++
        }
    }
    result := []SpanningTreeInstance{}
    for id, count := range instances {
        result = append(result, SpanningTreeInstance{InstanceID: id, BlockedPorts: count})
    }
    return result, nil
}