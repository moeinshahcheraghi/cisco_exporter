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
	maxConcurrent := 5
	sem := make(chan struct{}, maxConcurrent)
	
	wg := &sync.WaitGroup{}
	wg.Add(len(c.devices))
	
	for _, d := range c.devices {
		sem <- struct{}{} // Acquire
		go func(device *connector.Device) {
			defer func() {
				<-sem // Release
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

	collectors := c.collectors.collectorsForDevice(device)
	
	numWorkers := 3 
	collectorChan := make(chan collector.RPCCollector, len(collectors))
	
	for _, col := range collectors {
		collectorChan <- col
	}
	close(collectorChan)

	wg := &sync.WaitGroup{}
	
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for col := range collectorChan {
				select {
				case <-ctx.Done():
					log.Warnf("Context cancelled for %s", device.Host)
					return
				default:
					c.collectFromCollector(ctx, client, col, ch, l)
				}
			}
		}()
	}

	wg.Wait()
	
	client.ClearCache()
}

func (c *ciscoCollector) collectFromCollector(
	ctx context.Context,
	client *rpc.Client,
	col collector.RPCCollector,
	ch chan<- prometheus.Metric,
	labelValues []string,
) {
	ct := time.Now()
	
	resultChan := make(chan error, 1)
	metricsChan := make(chan prometheus.Metric, 100) 
	
	go func() {
		metrics := []prometheus.Metric{}
		for m := range metricsChan {
			metrics = append(metrics, m)
		}
		
		for _, m := range metrics {
			select {
			case <-ctx.Done():
				return
			default:
				ch <- m
			}
		}
	}()
	
	go func() {
		err := col.Collect(client, metricsChan, labelValues)
		close(metricsChan)
		resultChan <- err
	}()

	select {
	case <-ctx.Done():
		log.Warnf("Collector %s timed out for %s", col.Name(), labelValues[0])
		return
	case err := <-resultChan:
		if err != nil && err.Error() != "EOF" {
			log.Errorln(col.Name() + ": " + err.Error())
		}
	}

	duration := time.Since(ct).Seconds()
	ch <- prometheus.MustNewConstMetric(
		scrapeCollectorDurationDesc,
		prometheus.GaugeValue,
		duration,
		append(labelValues, col.Name())...,
	)
	
	if cfg.Debug {
		log.Infof("Collector %s for %s completed in %.2fs", col.Name(), labelValues[0], duration)
	}
}