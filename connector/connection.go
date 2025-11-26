package connector

import (
	"bufio"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
	"log"
	"sync"
	"github.com/moeinshahcheraghi/cisco_exporter/config"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type SSHConnection struct {
	client       *ssh.Client
	Host         string
	stdin        io.WriteCloser
	stdout       io.Reader
	session      *ssh.Session
	batchSize    int
	clientConfig *ssh.ClientConfig
	Debug        bool
	
	// Circuit breaker state
	failureCount int
	lastFailure  time.Time
	mutex        sync.Mutex
	maxRetries   int
	backoffTime  time.Duration
}

func NewSSSHConnection(device *Device, cfg *config.Config) (*SSHConnection, error) {
	deviceConfig := device.DeviceConfig

	legacyCiphers := cfg.LegacyCiphers
	if deviceConfig.LegacyCiphers != nil {
		legacyCiphers = *deviceConfig.LegacyCiphers
	}

	batchSize := cfg.BatchSize
	if deviceConfig.BatchSize != nil {
		batchSize = *deviceConfig.BatchSize
	}

	timeout := cfg.Timeout
	if deviceConfig.Timeout != nil {
		timeout = *deviceConfig.Timeout
	}

	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeout) * time.Second,
	}
	if legacyCiphers {
		sshConfig.SetDefaults()
		sshConfig.Ciphers = append(sshConfig.Ciphers, "aes128-cbc", "3des-cbc")
	}

	device.Auth(sshConfig)

	c := &SSHConnection{
		Host:         device.Host + ":" + device.Port,
		batchSize:    batchSize,
		clientConfig: sshConfig,
		Debug:        cfg.Debug,
		maxRetries:   3,
		backoffTime:  time.Second,
	}

	err := c.Connect()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *SSHConnection) Connect() error {
	var err error
	c.client, err = ssh.Dial("tcp", c.Host, c.clientConfig)
	if err != nil {
		c.recordFailure()
		return err
	}

	session, err := c.client.NewSession()
	if err != nil {
		c.client.Conn.Close()
		c.recordFailure()
		return err
	}
	c.stdin, _ = session.StdinPipe()
	c.stdout, _ = session.StdoutPipe()
	modes := ssh.TerminalModes{
		ssh.ECHO:  0,
		ssh.OCRNL: 0,
	}
	session.RequestPty("vt100", 0, 2000, modes)
	session.Shell()
	c.session = session

	c.RunCommand("")
	c.RunCommand("terminal length 0")

	c.resetFailures()
	return nil
}

type result struct {
	output string
	err    error
}

// RunCommand with circuit breaker and adaptive timeout
func (c *SSHConnection) RunCommand(cmd string) (string, error) {
	if c.shouldCircuitBreak() {
		return "", errors.New("Circuit breaker open - too many failures")
	}

	buf := bufio.NewReader(c.stdout)
	io.WriteString(c.stdin, cmd+"\n")

	outputChan := make(chan result, 1)
	
	go func() {
		c.readln(outputChan, cmd, buf)
	}()

	timeout := c.getAdaptiveTimeout()
	
	select {
	case res := <-outputChan:
		if res.err != nil {
			c.recordFailure()
		} else {
			c.resetFailures()
		}
		return res.output, res.err
	case <-time.After(timeout):
		c.recordFailure()
		if c.Debug {
			log.Printf("Timeout (%v) reached for command '%s' on %s (failures: %d)\n", 
				timeout, cmd, c.Host, c.failureCount)
		}
		return "", errors.New("Timeout reached")
	}
}

func (c *SSHConnection) Close() {
	if c.client != nil && c.client.Conn != nil {
		c.client.Conn.Close()
	}
	if c.session != nil {
		c.session.Close()
	}
}

// Circuit breaker helpers
func (c *SSHConnection) recordFailure() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.failureCount++
	c.lastFailure = time.Now()
}

func (c *SSHConnection) resetFailures() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.failureCount = 0
}

func (c *SSHConnection) shouldCircuitBreak() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if c.failureCount >= c.maxRetries {
		if time.Since(c.lastFailure) < c.backoffTime*time.Duration(c.failureCount) {
			return true
		}
		c.failureCount = 0
	}
	return false
}

func (c *SSHConnection) getAdaptiveTimeout() time.Duration {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	baseTimeout := c.clientConfig.Timeout
	if c.failureCount > 0 {
		return baseTimeout * time.Duration(1+c.failureCount)
	}
	return baseTimeout
}

func loadPrivateKey(r io.Reader) (ssh.AuthMethod, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not read from reader")
	}

	key, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse private key")
	}

	return ssh.PublicKeys(key), nil
}

func (c *SSHConnection) readln(ch chan result, cmd string, r io.Reader) {
	re := regexp.MustCompile(`.+#\s?$`)
	buf := make([]byte, c.batchSize)
	loadStr := ""
	
	deadline := time.Now().Add(c.clientConfig.Timeout)
	
	for {
		if time.Now().After(deadline) {
			ch <- result{output: "", err: errors.New("Read deadline exceeded")}
			return
		}
		
		n, err := r.Read(buf)
		if err != nil {
			ch <- result{output: "", err: err}
			return
		}
		loadStr += string(buf[:n])
		
		if strings.Contains(loadStr, cmd) && re.MatchString(loadStr) {
			break
		}
	}
	
	loadStr = strings.Replace(loadStr, "\r", "", -1)
	ch <- result{output: loadStr, err: nil}
}