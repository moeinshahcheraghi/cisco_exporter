package facts

import (
    "log"
    "strings"

    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_facts_"

var (
    versionDesc           *prometheus.Desc
    memoryTotalDesc       *prometheus.Desc
    memoryUsedDesc        *prometheus.Desc
    memoryFreeDesc        *prometheus.Desc
    cpuOneMinuteDesc      *prometheus.Desc
    cpuFiveSecondsDesc    *prometheus.Desc
    cpuInterruptsDesc     *prometheus.Desc
    cpuFiveMinutesDesc    *prometheus.Desc
    topProcessCPUDesc     *prometheus.Desc
    topProcessMemoryDesc  *prometheus.Desc
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
    if err := c.CollectVersion(client, ch, labelValues); err != nil && client.Debug {
        log.Printf("CollectVersion for %s: %s\n", labelValues[0], err)
    }
    if err := c.CollectMemory(client, ch, labelValues); err != nil && client.Debug {
        log.Printf("CollectMemory for %s: %s\n", labelValues[0], err)
    }
    if err := c.CollectCPU(client, ch, labelValues); err != nil && client.Debug {
        log.Printf("CollectCPU for %s: %s\n", labelValues[0], err)
    }
    if err := c.CollectTopProcessCPU(client, ch, labelValues); err != nil && client.Debug {
        log.Printf("CollectTopProcessCPU for %s: %s\n", labelValues[0], err)
    }
    if err := c.CollectTopProcessMemory(client, ch, labelValues); err != nil && client.Debug {
        log.Printf("CollectTopProcessMemory for %s: %s\n", labelValues[0], err)
    }
    return nil
}

func (c *factsCollector) CollectVersion(client *rpc.Client, ch chan<- prometheus.Metric, labels []string) error {
    out, err := client.RunCommand("show version")
    if err != nil {
        return err
    }
    version := parseVersionString(out)
    ch <- prometheus.MustNewConstMetric(versionDesc, prometheus.GaugeValue, 1.0, append(labels, version)...)
    return nil
}

func (c *factsCollector) CollectMemory(client *rpc.Client, ch chan<- prometheus.Metric, labels []string) error {
    out, err := client.RunCommand("show processes memory")
    if err != nil {
        return err
    }
    facts := ParseMemory(out)
    for _, f := range facts {
        ch <- prometheus.MustNewConstMetric(memoryTotalDesc, prometheus.GaugeValue, f.Total, append(labels, f.Type)...)
        ch <- prometheus.MustNewConstMetric(memoryUsedDesc, prometheus.GaugeValue, f.Used, append(labels, f.Type)...)
        ch <- prometheus.MustNewConstMetric(memoryFreeDesc, prometheus.GaugeValue, f.Free, append(labels, f.Type)...)
    }
    return nil
}

func (c *factsCollector) CollectCPU(client *rpc.Client, ch chan<- prometheus.Metric, labels []string) error {
    out, err := client.RunCommand("show processes cpu")
    if err != nil {
        return err
    }
    fact := ParseCPU(out)
    ch <- prometheus.MustNewConstMetric(cpuFiveSecondsDesc, prometheus.GaugeValue, fact.FiveSeconds, labels...)
    ch <- prometheus.MustNewConstMetric(cpuInterruptsDesc, prometheus.GaugeValue, fact.Interrupts, labels...)
    ch <- prometheus.MustNewConstMetric(cpuOneMinuteDesc, prometheus.GaugeValue, fact.OneMinute, labels...)
    ch <- prometheus.MustNewConstMetric(cpuFiveMinutesDesc, prometheus.GaugeValue, fact.FiveMinutes, labels...)
    return nil
}

func (c *factsCollector) CollectTopProcessCPU(client *rpc.Client, ch chan<- prometheus.Metric, labels []string) error {
    out, err := client.RunCommand("show processes cpu sorted | exclude 0.00%")
    if err != nil {
        return err
    }
    process, err := ParseTopProcessCPU(client.OSType, out)
    if err != nil {
        return err
    }
    ch <- prometheus.MustNewConstMetric(topProcessCPUDesc, prometheus.GaugeValue, process.CPUUsage, append(labels, process.Name)...)
    return nil
}

func (c *factsCollector) CollectTopProcessMemory(client *rpc.Client, ch chan<- prometheus.Metric, labels []string) error {
    out, err := client.RunCommand("show processes memory")
    if err != nil {
        return err
    }
    process, err := ParseTopProcessMemory(client.OSType, out)
    if err != nil {
        return err
    }
    ch <- prometheus.MustNewConstMetric(topProcessMemoryDesc, prometheus.GaugeValue, process.MemoryUsage, append(labels, process.Name)...)
    return nil
}

func parseVersionString(output string) string {
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if strings.Contains(line, "Version") {
            return strings.TrimSpace(line)
        }
    }
    return "unknown"
}
