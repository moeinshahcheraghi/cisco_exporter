package config

import (
	"io"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Debug         bool            `yaml:"debug"`
	LegacyCiphers bool            `yaml:"legacy_ciphers,omitempty"`
	Timeout       int             `yaml:"timeout,omitempty"`
	BatchSize     int             `yaml:"batch_size,omitempty"`
	Username      string          `yaml:"username,omitempty"`
	Password      string          `yaml:"Password,omitempty"`
	KeyFile       string          `yaml:"key_file,omitempty"`
	Devices       []*DeviceConfig `yaml:"devices,omitempty"`
	Features      *FeatureConfig  `yaml:"features,omitempty"`
}

type DeviceConfig struct {
	Host          string         `yaml:"host"`
	Username      *string        `yaml:"username,omitempty"`
	Password      *string        `yaml:"password,omitempty"`
	KeyFile       *string        `yaml:"key_file,omitempty"`
	LegacyCiphers *bool          `yaml:"legacy_ciphers,omitempty"`
	Timeout       *int           `yaml:"timeout,omitempty"`
	BatchSize     *int           `yaml:"batch_size,omitempty"`
	Features      *FeatureConfig `yaml:"features,omitempty"`
}

type FeatureConfig struct {
	BGP         *bool `yaml:"bgp,omitempty"`
	Environment *bool `yaml:"environment,omitempty"`
	Facts       *bool `yaml:"facts,omitempty"`
	Interfaces  *bool `yaml:"interfaces,omitempty"`
	Optics      *bool `yaml:"optics,omitempty"`
	StackPort   *bool `yaml:"stack_port,omitempty"`
	TablesARP       *bool `yaml:"tables_arp,omitempty"`
	TablesMAC       *bool `yaml:"tables_mac,omitempty"`
	TablesRouteIPv4 *bool `yaml:"tables_route_ipv4,omitempty"`
	TablesRouteIPv6 *bool `yaml:"tables_route_ipv6,omitempty"`
}

func New() *Config {
	c := &Config{
		Features: &FeatureConfig{},
	}
	c.setDefaultValues()
	return c
}

func Load(reader io.Reader) (*Config, error) {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	c := New()
	err = yaml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	for _, d := range c.Devices {
		if d.Features == nil {
			continue
		}
		if d.Features.BGP == nil {
			d.Features.BGP = c.Features.BGP
		}
		if d.Features.Environment == nil {
			d.Features.Environment = c.Features.Environment
		}
		if d.Features.Facts == nil {
			d.Features.Facts = c.Features.Facts
		}
		if d.Features.Interfaces == nil {
			d.Features.Interfaces = c.Features.Interfaces
		}
		if d.Features.Optics == nil {
			d.Features.Optics = c.Features.Optics
		}
		if d.Features.StackPort == nil {
			d.Features.StackPort = c.Features.StackPort
		}
	}

	return c, nil
}

func (c *Config) setDefaultValues() {
	c.Debug = false
	c.LegacyCiphers = false
	c.Timeout = 5
	c.BatchSize = 10000

	f := c.Features
	bgp := true
	f.BGP = &bgp
	environment := true
	f.Environment = &environment
	facts := true
	f.Facts = &facts
	interfaces := true
	f.Interfaces = &interfaces
	optics := true
	f.Optics = &optics
	stackPort := true
	f.StackPort = &stackPort
	tablesARP := true
	c.Features.TablesARP = &tablesARP
	tablesMAC := true
	c.Features.TablesMAC = &tablesMAC
	tablesRouteIPv4 := true
	c.Features.TablesRouteIPv4 = &tablesRouteIPv4
	tablesRouteIPv6 := true
	c.Features.TablesRouteIPv6 = &tablesRouteIPv6
}

func (c *Config) DevicesFromTargets(sshHosts string) {
	targets := strings.Split(sshHosts, ",")

	c.Devices = make([]*DeviceConfig, len(targets))
	for i, target := range targets {
		c.Devices[i] = &DeviceConfig{
			Host: target,
		}
	}
}

func (c *Config) FeaturesForDevice(host string) *FeatureConfig {
	d := c.findDeviceConfig(host)
	if d != nil && d.Features != nil {
		return d.Features
	}
	return c.Features
}

func (c *Config) findDeviceConfig(host string) *DeviceConfig {
	for _, dc := range c.Devices {
		if dc.Host == host {
			return dc
		}
	}
	return nil
}