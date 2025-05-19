package facts

import (
    "log"
    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_facts_"

var (
    versionDesc        *prometheus.Desc
    memoryTotalDesc    *prometheus.Desc
    memoryUsedDesc     *prometheus.Desc
    memoryFreeDesc     *prometheus.Desc
    cpuOneMinuteDesc   *prometheus.Desc
    cpuFiveSecondsDesc *prometheus.Desc
    cpuInterruptsDesc  *prometheus.Desc
    cpuFiveMinutesDesc *prometheus.Desc
    topProcessCPUDesc  *prometheus.Desc 
    topProcessMemoryDesc *prometheus.Desc 
)

func init() {
    l := []string{"target"}
    versionDesc = prometheus.NewDesc(prefix+"version", "Running OS version", append(l, "version"), nil)
    memoryTotalDesc = prometheus.NewDesc(prefix+"memory_total", "Total memory", append(l, "type"), nil)
    memoryUsedDesc = prometheus.NewDesc(prefix+"memory_used", "Used memory", append(l, "type"), nil)
    memoryFreeDesc = prometheus.NewDesc(prefix+"memory_free", "Free memory", append(l, "type"), nil)
    cpuOneMinuteDesc = prometheus.NewDesc(prefix+"cpu_one_minute_percent", "CPU utilization for one minute", l, nil)
    cpuFiveSecondsDesc = prometheus.NewDesc(prefix+"cpu_five_seconds_percent", "CPU utilization for five seconds", l, nil)
    cpuInterruptsDesc = prometheus.NewDesc(prefix+"cpu_interrupt_percent", "Interrupt percentage", l, nil)
    cpuFiveMinutesDesc = prometheus.NewDesc(prefix+"cpu_five_minutes_percent", "CPU utilization for five minutes", l, nil)
    // Descriptor haye jadid baraye top process
    topProcessCPUDesc = prometheus.NewDesc(prefix+"top_process_cpu_percent", "CPU usage of the top process", append(l, "process_name"), nil)
    topProcessMemoryDesc = prometheus.NewDesc(prefix+"top_process_memory_bytes", "Memory usage of the top process", append(l, "process_name"), nil)
}

type factsCollector struct{}

func NewCollector() collector.RPCCollector {
    return &factsCollector{}
}

func (*factsCollector) Name() string {
    return "Facts"
}

func (*factsCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- versionDesc
    ch <- memoryTotalDesc
    ch <- memoryUsedDesc
    ch <- memoryFreeDesc
    ch <- cpuOneMinuteDesc
    ch <- cpuFiveSecondsDesc
    ch <- cpuInterruptsDesc
    ch <- cpuFiveMinutesDesc
    ch <- topProcessCPUDesc 
    ch <- topProcessMemoryDesc
}

func (c *factsCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
    err := c.CollectVersion(client, ch, labelValues)
    if client.Debug && err != nil {
        log.Printf("CollectVersion for %s: %s\n", labelValues[0], err.Error())
    }
    err = c.CollectMemory(client, ch, labelValues)
    if client.Debug && err != nil {
        log.Printf("CollectMemory for %s: %s\n", labelValues[0], err.Error())
    }
    err = c.CollectCPU(client, ch, labelValues)
    if client.Debug && err != nil {
        log.Printf("CollectCPU for %s: %s\n", labelValues[0], err.Error())
    }
    err = c.CollectTopProcessCPU(client, ch, labelValues)
    if client.Debug && err != nil {
        log.Printf("CollectTopProcessCPU for %s: %s\n", labelValues[0], err.Error())
    }
    err = c.CollectTopProcessMemory(client, ch, labelValues)
    if client.Debug && err != nil {
        log.Printf("CollectTopProcessMemory for %s: %s\n", labelValues[0], err.Error())
    }
    return nil
}