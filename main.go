package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/moeinshahcheraghi/cisco_exporter/config"
	"github.com/moeinshahcheraghi/cisco_exporter/connector"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const version string = "0.2"

var (
	showVersion         = flag.Bool("version", false, "Print version information.")
	listenAddress       = flag.String("web.listen-address", ":9362", "Address on which to expose metrics and web interface.")
	metricsPath         = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	sshHosts            = flag.String("ssh.targets", "", "Comma-separated list of SSH targets")
	sshUsername         = flag.String("ssh.user", "cisco_exporter", "Username to use for SSH connection")
	sshPassword         = flag.String("ssh.password", "", "Password to use for SSH connection")
	sshKeyFile          = flag.String("ssh.keyfile", "", "Key file to use for SSH connection")
	sshTimeout          = flag.Int("ssh.timeout", 5, "Timeout to use for SSH connection")
	sshBatchSize        = flag.Int("ssh.batch-size", 10000, "The SSH response batch size")
	debug               = flag.Bool("debug", false, "Show verbose debug output in log")
	legacyCiphers       = flag.Bool("legacy.ciphers", false, "Allow legacy CBC ciphers")
	bgpEnabled          = flag.Bool("bgp.enabled", true, "Scrape BGP metrics")
	environmentEnabled  = flag.Bool("environment.enabled", true, "Scrape environment metrics")
	factsEnabled        = flag.Bool("facts.enabled", true, "Scrape system facts (CPU/mem/version)")
	interfacesEnabled   = flag.Bool("interfaces.enabled", true, "Scrape interface metrics")
	opticsEnabled       = flag.Bool("optics.enabled", true, "Scrape optic metrics")
	stackportEnabled    = flag.Bool("stackport.enabled", true, "Scrape stack port metrics")
	etherchannelEnabled = flag.Bool("etherchannel.enabled", true, "Scrape EtherChannel metrics")
	slottempEnabled     = flag.Bool("slottemp.enabled", true, "Scrape slot temperature metrics")
	loginfailuresEnabled = flag.Bool("loginfailures.enabled", true, "Scrape login failure metrics")
	configlogEnabled    = flag.Bool("configlog.enabled", true, "Scrape config log metrics")
	spanningtreeEnabled = flag.Bool("spanningtree.enabled", true, "Scrape spanning tree metrics")
	processesEnabled    = flag.Bool("processes.enabled", true, "Scrape CPU/memory process metrics")
	arpEnabled          = flag.Bool("arp.enabled", true, "Scrape ARP metrics")
	cefEnabled          = flag.Bool("cef.enabled", true, "Scrape CEF drops metrics")
	configFile          = flag.String("config.file", "", "Path to config file (YAML)")

	cfg     *config.Config
	devices []*connector.Device
)

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	err := initialize()
	if err != nil {
		log.Fatalf("Error during initialization: %v", err)
	}

	startServer()
}

func initialize() error {
	c, err := loadConfig()
	if err != nil {
		return err
	}

	devices, err = devicesForConfig(c)
	if err != nil {
		return err
	}
	cfg = c

	return nil
}

func loadConfig() (*config.Config, error) {
	if *configFile == "" {
		log.Infoln("Loading config from flags")
		return loadConfigFromFlags(), nil
	}

	log.Infof("Loading config from file: %s", *configFile)
	b, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}
	return config.Load(bytes.NewReader(b))
}

func loadConfigFromFlags() *config.Config {
	c := config.New()

	c.Debug = *debug
	c.LegacyCiphers = *legacyCiphers
	c.Timeout = *sshTimeout
	c.BatchSize = *sshBatchSize
	c.Username = *sshUsername
	c.Password = *sshPassword
	c.KeyFile = *sshKeyFile

	c.DevicesFromTargets(*sshHosts)

	f := c.Features
	f.BGP = bgpEnabled
	f.Environment = environmentEnabled
	f.Facts = factsEnabled
	f.Interfaces = interfacesEnabled
	f.Optics = opticsEnabled
	f.StackPort = stackportEnabled
	f.EtherChannel = etherchannelEnabled
	f.SlotTemp = slottempEnabled
	f.LoginFailures = loginfailuresEnabled
	f.ConfigLog = configlogEnabled
	f.SpanningTree = spanningtreeEnabled
	f.Processes = processesEnabled
	f.ARP = arpEnabled
	f.CEF = cefEnabled

	return c
}

func printVersion() {
	fmt.Println("Cisco Exporter")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author: Martin Poppen and contributors")
	fmt.Println("Exports metrics from Cisco IOS/NX-OS/IOS-XE devices")
}

func startServer() {
	log.Infof("Starting Cisco exporter (Version %s)", version)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Cisco Exporter</title></head>
			<body>
			<h1>Cisco Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	http.Handle(*metricsPath, promhttp.Handler())

	log.Infof("Listening on %s for metrics path %s", *listenAddress, *metricsPath)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
