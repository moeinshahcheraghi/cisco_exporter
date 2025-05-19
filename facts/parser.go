package facts

import (
    "errors"
    "regexp"
    "strings"

    "github.com/moeinshahcheraghi/cisco_exporter/util"
)
func ParseTopProcessCPU(ostype string, output string) (Process, error) {
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
    return Process{}, errors.New("no active process found")
}

func ParseTopProcessMemory(ostype string, output string) (Process, error) {
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
        return Process{}, errors.New("no process with memory usage found")
    }
    return topProcess, nil
}