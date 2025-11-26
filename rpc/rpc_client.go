package rpc

import (
	"errors"
	"fmt"
	"strings"
	"log"
	"sync"
	"time"
	"github.com/moeinshahcheraghi/cisco_exporter/connector"
)

const (
	IOSXE string = "IOSXE"
	NXOS  string = "NXOS"
	IOS   string = "IOS"
)

type cacheEntry struct {
	value     string
	timestamp time.Time
}

// Client sends commands to a Cisco device
type Client struct {
	conn       *connector.SSHConnection
	Debug      bool
	OSType     string
	cache      map[string]*cacheEntry
	cacheMutex sync.RWMutex
	cacheTTL   time.Duration
}

// NewClient creates a new client connection
func NewClient(ssh *connector.SSHConnection, debug bool) *Client {
	return &Client{
		conn:     ssh,
		Debug:    debug,
		cache:    make(map[string]*cacheEntry),
		cacheTTL: 30 * time.Second,
	}
}

// Identify tries to identify the OS running on a Cisco device
func (c *Client) Identify() error {
	output, err := c.RunCommand("show version")
	if err != nil {
		return err
	}
	switch {
	case strings.Contains(output, "IOS XE"):
		c.OSType = IOSXE
	case strings.Contains(output, "NX-OS"):
		c.OSType = NXOS
	case strings.Contains(output, "IOS Software"):
		c.OSType = IOS
	default:
		return errors.New("Unknown OS")
	}
	if c.Debug {
		log.Printf("Host %s identified as: %s\n", c.conn.Host, c.OSType)
	}
	return nil
}

// RunCommand runs a command with improved caching and timeout handling
func (c *Client) RunCommand(cmd string) (string, error) {
	c.cacheMutex.RLock()
	if entry, ok := c.cache[cmd]; ok {
		if time.Since(entry.timestamp) < c.cacheTTL {
			c.cacheMutex.RUnlock()
			if c.Debug {
				log.Printf("Cache hit for command '%s' on %s (age: %v)\n", 
					cmd, c.conn.Host, time.Since(entry.timestamp))
			}
			return entry.value, nil
		}
	}
	c.cacheMutex.RUnlock()

	if c.Debug {
		log.Printf("Running command on %s: %s\n", c.conn.Host, cmd)
	}

	start := time.Now()
	output, err := c.conn.RunCommand(fmt.Sprintf("%s", cmd))
	duration := time.Since(start)

	if err == nil {
		c.cacheMutex.Lock()
		c.cache[cmd] = &cacheEntry{
			value:     output,
			timestamp: time.Now(),
		}
		c.cacheMutex.Unlock()

		if c.Debug {
			log.Printf("Command '%s' on %s succeeded in %v. Output cached.\n", 
				cmd, c.conn.Host, duration)
		}
	} else {
		if c.Debug {
			log.Printf("Command '%s' on %s failed after %v: %s\n", 
				cmd, c.conn.Host, duration, err.Error())
		}
	}

	return output, err
}

func (c *Client) ClearCache() {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	
	now := time.Now()
	for cmd, entry := range c.cache {
		if now.Sub(entry.timestamp) > c.cacheTTL {
			delete(c.cache, cmd)
		}
	}
}