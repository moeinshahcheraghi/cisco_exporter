package facts

type MemoryFact struct {
    Used  float64
    Total float64
}

type CPUFact struct {
    FiveSeconds float64
    OneMinute   float64
    FiveMinutes float64
}

type Process struct {
    Name        string
    CPUUsage    float64
    MemoryUsage float64
}