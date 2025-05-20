package facts

import (
    "errors"
    "regexp"
    "strconv"
    "strings"
)

// ParseMemory parses the output of "show processes memory"
func ParseMemory(output string) []MemoryFact {
    re := regexp.MustCompile(`Processor Pool Total:\s+(\d+)\s+Used:\s+(\d+)`)
    matches := re.FindStringSubmatch(output)
    if matches != nil {
        total, _ := strconv.ParseFloat(matches[1], 64)
        used, _ := strconv.ParseFloat(matches[2], 64)
        return []MemoryFact{{Total: total, Used: used}}
    }
    return []MemoryFact{}
}

// ParseCPU parses the output of "show processes cpu"
func ParseCPU(output string) CPUFact {
    re := regexp.MustCompile(`CPU utilization for five seconds: (\d+)%.*?one minute: (\d+)%.*?five minutes: (\d+)%`)
    matches := re.FindStringSubmatch(output)
    if matches != nil {
        fiveSeconds, _ := strconv.ParseFloat(matches[1], 64)
        oneMinute, _ := strconv.ParseFloat(matches[2], 64)
        fiveMinutes, _ := strconv.ParseFloat(matches[3], 64)
        return CPUFact{FiveSeconds: fiveSeconds, OneMinute: oneMinute, FiveMinutes: fiveMinutes}
    }
    return CPUFact{}
}

// ParseTopProcessCPU parses the output of "show processes cpu sorted | exclude 0.00%"
func ParseTopProcessCPU(ostype string, output string) (Process, error) {
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if strings.Contains(line, "PID") {
            continue
        }
        fields := strings.Fields(line)
        if len(fields) >= 9 {
            name := fields[8]
            cpuUsage, _ := strconv.ParseFloat(fields[5], 64)
            return Process{Name: name, CPUUsage: cpuUsage}, nil
        }
    }
    return Process{}, errors.New("no top process found")
}

// ParseTopProcessMemory parses the output of "show processes memory"
func ParseTopProcessMemory(ostype string, output string) (Process, error) {
    re := regexp.MustCompile(`(\d+)\s+\d+\s+\d+\s+(\d+)\s+\S+`)
    lines := strings.Split(output, "\n")
    var topProcess Process
    maxMemory := float64(0)
    for _, line := range lines {
        matches := re.FindStringSubmatch(line)
        if matches != nil {
            memory, _ := strconv.ParseFloat(matches[2], 64)
            if memory > maxMemory {
                maxMemory = memory
                topProcess = Process{Name: matches[1], MemoryUsage: memory}
            }
        }
    }
    if maxMemory == 0 {
        return Process{}, errors.New("no process with memory usage found")
    }
    return topProcess, nil
}