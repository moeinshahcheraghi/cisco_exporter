package facts

import (
    "errors"
    "regexp"
    "strings"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/moeinshahcheraghi/cisco_exporter/util"
)

func ParseTopProcessCPU(ostype string, output string) (Process, error) {
    if ostype != rpc.IOSXE && ostype != rpc.IOS {
        return Process{}, errors.New("'show processes cpu sorted | exclude 0.00%' is not implemented for " + ostype)
    }
    processRegexp := regexp.MustCompile(`^\s*(\d+)\s+\d+\s+\d+\s+\d+\s+(\d+\.\d+)%\s+\d+\.\d+%\s+\d+\.\d+%\s+\d+\s+(\S+)`)
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if matches := processRegexp.FindStringSubmatch(line); matches != nil {
            return Process{
                Name:     matches[3],
                CPUUsage: util.Str2float64(matches[2]),
            }, nil
        }
    }
    return Process{}, errors.New("No active process found")
}

func ParseTopProcessMemory(ostype string, output string) (Process, error) {
    if ostype != rpc.IOSXE && ostype != rpc.IOS {
        return Process{}, errors.New("'show processes memory' is not implemented for " + ostype)
    }
    processRegexp := regexp.MustCompile(`^\s*(\d+)\s+\d+\s+\d+\s+\d+\s+(\d+)\s+\S+`)
    lines := strings.Split(output, "\n")
    var topProcess Process
    maxMemory := float64(0)
    for _, line := range lines {
        if matches := processRegexp.FindStringSubmatch(line); matches != nil {
            memory := util.Str2float64(matches[2])
            if memory > maxMemory {
                maxMemory = memory
                topProcess = Process{
                    Name:        matches[1], 
                    MemoryUsage: memory,
                }
            }
        }
    }
    if maxMemory == 0 {
        return Process{}, errors.New("No process with memory usage found")
    }
    return topProcess, nil
}