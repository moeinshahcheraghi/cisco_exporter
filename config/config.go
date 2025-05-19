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
    BGP          *bool `yaml:"bgp,omitempty"`
    Environment  *bool `yaml:"environment,omitempty"`
    Facts        *bool `yaml:"facts,omitempty"`
    Interfaces   *bool `yaml:"interfaces,omitempty"`
    Optics       *bool `yaml:"optics,omitempty"`
    StackPort    *bool `yaml:"stack_port,omitempty"`
    Etherchannel *bool `yaml:"etherchannel,omitempty"`
    Login        *bool `yaml:"login,omitempty"`
    SpanningTree *bool `yaml:"spanningtree,omitempty"`
    ARP          *bool `yaml:"arp,omitempty"`
}

func New() *Config {
    c := &Config{Features: &FeatureConfig{}}
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
        if d.Features.BGP == nil { d.Features.BGP = c.Features.BGP }
        if d.Features.Environment == nil { d.Features.Environment = c.Features.Environment }
        if d.Features.Facts == nil { d.Features.Facts = c.Features.Facts }
        if d.Features.Interfaces == nil { d.Features.Interfaces = c.Features.Interfaces }
        if d.Features.Optics == nil { d.Features.Optics = c.Features.Optics }
        if d.Features.StackPort == nil { d.Features.StackPort = c.Features.StackPort }
        if d.Features.Etherchannel == nil { d.Features.Etherchannel = c.Features.Etherchannel }
        if d.Features.Login == nil { d.Features.Login = c.Features.Login }
        if d.Features.SpanningTree == nil { d.Features.SpanningTree = c.Features.SpanningTree }
        if d.Features.ARP == nil { d.Features.ARP = c.Features.ARP }
    }
    return c, nil
}

func (c *Config) setDefaultValues() {
    c.Debug = false
    c.LegacyCiphers = false
    c.Timeout = 5
    c.BatchSize = 10000

    f := c.Features
    enabled := true
    f.BGP = &enabled
    f.Environment = &enabled
    f.Facts = &enabled
    f.Interfaces = &enabled
    f.Optics = &enabled
    f.StackPort = &enabled
    f.Etherchannel = &enabled
    f.Login = &enabled
    f.SpanningTree = &enabled
    f.ARP = &enabled
}

func (c *Config) DevicesFromTargets(sshHosts string) {
    targets := strings.Split(sshHosts, ",")
    c.Devices = make([]*DeviceConfig, len(targets))
    for i, target := range targets {
        c.Devices[i] = &DeviceConfig{Host: target}
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