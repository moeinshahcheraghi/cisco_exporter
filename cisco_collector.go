package main

import (
    "sync"
    "time"
    "github.com/moeinshahcheraghi/cisco_exporter/connector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
    log "github.com/sirupsen/logrus"
)

const prefix = "cisco_"

var (
    scrapeCollectorDurationDesc *prometheus.Desc
    scrapeDurationDesc          *prometheus.Desc
    upDesc                      *prometheus.Desc
)

func init() {
    upDesc = prometheus.NewDesc(prefix+"up", "Scrape of target was successful", []string{"target"}, nil)
    scrapeDurationDesc = prometheus.NewDesc(prefix+"collector_duration_seconds", "Duration of a collector scrape for one target", []string{"target"}, nil)
    scrapeCollectorDurationDesc = prometheus.NewDesc(prefix+"collect_duration_seconds", "Duration of a scrape by collector and target", []string{"target", "collector"}, nil)
}

type ciscoCollector struct {
    devices    []*connector.Device
    collectors *collectors
}

func newCiscoCollector(devices []*connector.Device) *ciscoCollector {
    return &ciscoCollector{
        devices:    devices,
        collectors: collectorsForDevices(devices, cfg),
    }
}

func (c *ciscoCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- upDesc
    ch <- scrapeDurationDesc
    ch <- scrapeCollectorDurationDesc
    for _, col := range c.collectors.allEnabledCollectors() {
        col.Describe(ch)
    }
}

func (c *ciscoCollector) Collect(ch chan<- prometheus.Metric) {
    var wg sync.WaitGroup
    wg.Add(len(c.devices))
    for _, d := range c.devices {
        go func(device *connector.Device) {
            defer wg.Done()
            c.collectForHost(device, ch)
        }(d)
    }
    wg.Wait()
}

func (c *ciscoCollector) collectForHost(device *connector.Device, ch chan<- prometheus.Metric) {
    l := []string{device.Host}
    t := time.Now()
    defer func() {
        ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(t).Seconds(), l...)
    }()

    conn, err := connector.NewSSSHConnection(device, cfg)
    if err != nil {
        log.Errorln(err)
        ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0, l...)
        return
    }
    defer conn.Close()

    ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1, l...)
    client := rpc.NewClient(conn, cfg.Debug)
    err = client.Identify()
    if err != nil {
        log.Errorln(device.Host + ": " + err.Error())
        return
    }

    var collectorWg sync.WaitGroup
    collectorChan := make(chan struct {
        collectorName string
        duration      float64
        err           error
    }, len(c.collectors.collectorsForDevice(device)))

    for _, col := range c.collectors.collectorsForDevice(device) {
        collectorWg.Add(1)
        go func(col collector.RPCCollector) {
            defer collectorWg.Done()
            ct := time.Now()
            err := col.Collect(client, ch, l)
            collectorChan <- struct {
                collectorName string
                duration      float64
                err           error
            }{col.Name(), time.Since(ct).Seconds(), err}
        }(col)
    }

    go func() {
        collectorWg.Wait()
        close(collectorChan)
    }()

    for result := range collectorChan {
        if result.err != nil && result.err.Error() != "EOF" {
            log.Errorln(result.collectorName + ": " + result.err.Error())
        }
        ch <- prometheus.MustNewConstMetric(scrapeCollectorDurationDesc, prometheus.GaugeValue, result.duration, append(l, result.collectorName)...)
    }
}