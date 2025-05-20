package facts

import (
    "regexp"
    "strconv"
)

// MemoryFact represents memory usage facts
type MemoryFact struct {
    Type  string
    Total float64
    Used  float64
    Free  float64
}

// CPUFact represents CPU usage facts
type CPUFact struct {
    Usage float64
}

// ParseMemory parses the output of a memory-related command
func ParseMemory(output string) []MemoryFact {
    re := regexp.MustCompile(`Processor Pool Total:\s+(\d+)\s+Used:\s+(\d+)\s+Free:\s+(\d+)`)
    matches := re.FindStringSubmatch(output)
    if matches != nil {
        total, _ := strconv.ParseFloat(matches[1], 64)
        used, _ := strconv.ParseFloat(matches[2], 64)
        free, _ := strconv.ParseFloat(matches[3], 64)
        return []MemoryFact{{Type: "Processor", Total: total, Used: used, Free: free}}
    }
    return []MemoryFact{}
}

// ParseCPU parses the output of a CPU-related command
func ParseCPU(output string) CPUFact {
    re := regexp.MustCompile(`CPU utilization.*? (\d+)%`)
    matches := re.FindStringSubmatch(output)
    if matches != nil {
        usage, _ := strconv.ParseFloat(matches[1], 64)
        return CPUFact{Usage: usage}
    }
    return CPUFact{}
}