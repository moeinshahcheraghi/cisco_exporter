package main

import (
	"github.com/moeinshahcheraghi/cisco_exporter/bgp"
	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/config"
	"github.com/moeinshahcheraghi/cisco_exporter/connector"
	"github.com/moeinshahcheraghi/cisco_exporter/environment"
	"github.com/moeinshahcheraghi/cisco_exporter/facts"
	"github.com/moeinshahcheraghi/cisco_exporter/interfaces"
	"github.com/moeinshahcheraghi/cisco_exporter/optics"
	"github.com/moeinshahcheraghi/cisco_exporter/stackport"
	"github.com/moeinshahcheraghi/cisco_exporter/tables"
	"github.com/moeinshahcheraghi/cisco_exporter/uptime"
    "github.com/moeinshahcheraghi/cisco_exporter/stp"
    "github.com/moeinshahcheraghi/cisco_exporter/vlan"
    "github.com/moeinshahcheraghi/cisco_exporter/qos"
    "github.com/moeinshahcheraghi/cisco_exporter/acl"

)

type collectors struct {
	collectors map[string]collector.RPCCollector
	devices    map[string][]collector.RPCCollector
	cfg        *config.Config
}

func collectorsForDevices(devices []*connector.Device, cfg *config.Config) *collectors {
	c := &collectors{
		collectors: make(map[string]collector.RPCCollector),
		devices:    make(map[string][]collector.RPCCollector),
		cfg:        cfg,
	}

	for _, d := range devices {
		c.initCollectorsForDevice(d)
	}

	return c
}

func (c *collectors) initCollectorsForDevice(device *connector.Device) {
	f := c.cfg.FeaturesForDevice(device.Host)

	c.devices[device.Host] = make([]collector.RPCCollector, 0)
	c.addCollectorIfEnabledForDevice(device, "bgp", f.BGP, bgp.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "environment", f.Environment, environment.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "facts", f.Facts, facts.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "interfaces", f.Interfaces, interfaces.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "optics", f.Optics, optics.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "stackport", f.StackPort, stackport.NewCollector) 
	c.addCollectorIfEnabledForDevice(device, "tablesARP", f.TablesARP, tables.NewARPCollector)
	c.addCollectorIfEnabledForDevice(device, "tablesMAC", f.TablesMAC, tables.NewMACCollector)
	c.addCollectorIfEnabledForDevice(device, "tablesRouteIPv4", f.TablesRouteIPv4, tables.NewRouteIPv4Collector)
	c.addCollectorIfEnabledForDevice(device, "tablesRouteIPv6", f.TablesRouteIPv6, tables.NewRouteIPv6Collector)
	c.addCollectorIfEnabledForDevice(device, "uptime", f.Uptime, uptime.NewCollector)
    c.addCollectorIfEnabledForDevice(device, "stp", f.STP, stp.NewCollector)
    c.addCollectorIfEnabledForDevice(device, "vlan", f.VLAN, vlan.NewCollector)
    c.addCollectorIfEnabledForDevice(device, "qos", f.QoS, qos.NewCollector)
    c.addCollectorIfEnabledForDevice(device, "acl", f.ACL, acl.NewCollector)
}

func (c *collectors) addCollectorIfEnabledForDevice(device *connector.Device, key string, enabled *bool, newCollector func() collector.RPCCollector) {
	if !*enabled {
		return
	}

	col, found := c.collectors[key]
	if !found {
		col = newCollector()
		c.collectors[key] = col
	}

	c.devices[device.Host] = append(c.devices[device.Host], col)
}

func (c *collectors) allEnabledCollectors() []collector.RPCCollector {
	collectors := make([]collector.RPCCollector, len(c.collectors))

	i := 0
	for _, collector := range c.collectors {
		collectors[i] = collector
		i++
	}

	return collectors
}

func (c *collectors) collectorsForDevice(device *connector.Device) []collector.RPCCollector {
	cols, found := c.devices[device.Host]
	if !found {
		return []collector.RPCCollector{}
	}

	return cols
}
