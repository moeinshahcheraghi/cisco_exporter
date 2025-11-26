package main

import (
    "context"
    "sync"
    "time"

    "github.com/moeinshahcheraghi/cisco_exporter/collector"
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
    wg := &sync.WaitGroup{}
    semaphore := make(chan struct{}, 3)

    wg.Add(len(c.devices))
    for _, d := range c.devices {
        go func(device *connector.Device) {
            semaphore <- struct{}{} 
            defer func() {
                <-semaphore 
                wg.Done()
            }()
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

    ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
    defer cancel()

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

    collectors := c.prioritizeCollectors(c.collectors.collectorsForDevice(device))

    collectorWg := &sync.WaitGroup{}
    collectorSem := make(chan struct{}, 2) 
    resultChan := make(chan prometheus.Metric, 100)

    go func() {
        for metric := range resultChan {
            select {
            case <-ctx.Done():
                return
            default:
                ch <- metric
            }
        }
    }()

    for _, col := range collectors {
        select {
        case <-ctx.Done():
            log.Warnf("%s: collection timeout reached", device.Host)
            goto cleanup
        default:
        }

        collectorWg.Add(1)
        go func(col collector.RPCCollector) {
            defer collectorWg.Done()

            collectorSem <- struct{}{}
            defer func() { <-collectorSem }()

            ct := time.Now()
            
            localCh := make(chan prometheus.Metric, 50)
            errCh := make(chan error, 1)

            go func() {
                errCh <- col.Collect(client, localCh, l)
                close(localCh)
            }()

            go func() {
                for metric := range localCh {
                    resultChan <- metric
                }
            }()

            select {
            case err := <-errCh:
                if err != nil && err.Error() != "EOF" {
                    log.Warnln(col.Name() + " on " + device.Host + ": " + err.Error())
                }
            case <-time.After(30 * time.Second):
                log.Warnf("%s collector timeout on %s", col.Name(), device.Host)
            case <-ctx.Done():
                log.Warnf("%s collector cancelled on %s", col.Name(), device.Host)
            }

            duration := time.Since(ct).Seconds()
            resultChan <- prometheus.MustNewConstMetric(
                scrapeCollectorDurationDesc,
                prometheus.GaugeValue,
                duration,
                append(l, col.Name())...,
            )
        }(col)
    }

cleanup:
    collectorWg.Wait()
    close(resultChan)
}

func (c *ciscoCollector) prioritizeCollectors(collectors []collector.RPCCollector) []collector.RPCCollector {
    lightWeight := []string{"Facts", "Uptime", "STP", "VLAN", "TablesARP", "TablesMAC"}
    heavyWeight := []string{"Interfaces", "BGP", "Environment", "Optics", "TablesRouteIPv4", "TablesRouteIPv6"}

    result := make([]collector.RPCCollector, 0, len(collectors))
    
    for _, name := range lightWeight {
        for _, col := range collectors {
            if col.Name() == name {
                result = append(result, col)
                break
            }
        }
    }
    
    for _, name := range heavyWeight {
        for _, col := range collectors {
            if col.Name() == name {
                result = append(result, col)
                break
            }
        }
    }
    
    for _, col := range collectors {
        found := false
        for _, added := range result {
            if added.Name() == col.Name() {
                found = true
                break
            }
        }
        if !found {
            result = append(result, col)
        }
    }
    
    return result
}