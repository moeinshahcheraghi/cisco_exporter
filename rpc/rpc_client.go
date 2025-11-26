package rpc

import (
	"errors"
	"fmt"
	"strings"
	"log"
	"github.com/moeinshahcheraghi/cisco_exporter/connector"
)

const (
	IOSXE string = "IOSXE"
	NXOS  string = "NXOS"
	IOS   string = "IOS"
	
)

// Client sends commands to a Cisco device
type Client struct {
    conn   *connector.SSHConnection
    Debug  bool
    OSType string
    cache  map[string]string
}

// NewClient creates a new client connection
func NewClient(ssh *connector.SSHConnection, debug bool) *Client {
    return &Client{
        conn:  ssh,
        Debug: debug,
        cache: make(map[string]string),
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

// RunCommand runs a command on a Cisco device with enhanced logging
func (c *Client) RunCommand(cmd string) (string, error) {
    if output, ok := c.cache[cmd]; ok {
        if c.Debug {
            log.Printf("Cache hit for command '%s' on %s\n", cmd, c.conn.Host)
        }
        return output, nil
    }
    if c.Debug {
        log.Printf("Running command on %s: %s\n", c.conn.Host, cmd)
    }
    output, err := c.conn.RunCommand(fmt.Sprintf("%s", cmd))
    if err == nil {
        c.cache[cmd] = output
    }
    if c.Debug {
        if err != nil {
            log.Printf("Command '%s' on %s failed: %s\n", cmd, c.conn.Host, err.Error())
        } else {
            log.Printf("Command '%s' on %s succeeded. Output cached.\n", cmd, c.conn.Host)
        }
    }
    return output, err
}