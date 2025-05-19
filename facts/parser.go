package facts

import (
    "errors"
    "regexp"
    "strconv"
    "strings"

    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/moeinshahcheraghi/cisco_exporter/util"
)

func ParseMemory(output string) []MemoryFact {
    // Example matching: Processor  12345678   8765432  3587231
    memoryRegexp := regexp.MustCompile(`(?i)^(Processor|I/O)\s+(\d+)\s+(\d+)\s+(\d+)`)
    facts := []MemoryFact{}
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if matches := memoryRegexp.FindStringSubmatch(line); matches != nil {
            total := util.Str2float64(matches[2])
            used := util.Str2float64(matches[3])
            free := util.Str2float64(matches[4])
            facts = append(facts, MemoryFact{
                Type:  matches[1],
                Total: total,
                Used:  used,
                Free:  free,
            })
        }
    }
    return facts
}

func ParseCPU(output string) CPUFact {
    // Example: CPU utilization for five seconds: 5%/0%; one minute: 5%; five minutes: 6%
    cpuRegexp := regexp.MustCompile(`CPU utilization for five seconds: (\d+)%/(\d+)%.*?one minute: (\d+)%.*?five minutes: (\d+)%`)
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if matches := cpuRegexp.FindStringSubmatch(line); matches != nil {
            return CPUFact{
                FiveSeconds: util.Str2float64(matches[1]),
                Interrupts:  util.Str2float64(matches[2]),
                OneMinute:   util.Str2float64(matches[3]),
                FiveMinutes: util.Str2float64(matches[4]),
            }
        }
    }
    return CPUFact{} // Return empty on failure (caller should handle)
}
