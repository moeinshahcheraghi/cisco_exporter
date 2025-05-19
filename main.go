package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/moeinshahcheraghi/cisco_exporter/collector"
	"github.com/moeinshahcheraghi/cisco_exporter/config"
	"github.com/moeinshahcheraghi/cisco_exporter/connector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const version string = "0.2"

var (
	showVersion        = flag.Bool("version", false, "Print version information.")
	listenAddress      = flag.String("web.listen-address", ":9362", "Address on which to expose metrics and web interface.")
	metricsPath        = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	sshHosts           = flag.String("ssh.targets", "", "SSH Hosts to scrape")
	sshUsername        = flag.String("ssh.user", "cisco_exporter", "Username to use for SSH connection")
	sshPassword        = flag.String("ssh.password", "", "Password to use for SSH connection")
	sshKeyFile         = flag.String("ssh.keyfile", "", "Key file to use for SSH connection")
	sshTimeout         = flag.Int("ssh.timeout", 5, "Timeout to use for SSH connection")
	sshBatchSize       = flag.Int("ssh.batch-size", 10000, "The SSH response batch size")
	debug              = flag.Bool("debug", false, "Show verbose debug output in log")
	legacyCiphers      = flag.Bool("legacy.ciphers", false, "Allow legacy CBC ciphers")
	bgpEnabled         = flag.Bool("bgp.enabled", true, "Scrape bgp metrics")
	environmentEnabled = flag.Bool("environment.enabled", true, "Scrape environment metrics")
	factsEnabled       = flag.Bool("facts.enabled", true, "Scrape system metrics")
	interfacesEnabled  = flag.Bool("interfaces.enabled", true, "Scrape interface metrics")
	opticsEnabled      = flag.Bool("optics.enabled", true, "Scrape optic metrics")
	configFile         = flag.String("config.file", "", "Path to config file")
<<<<<<< HEAD
	stackportEnabled = flag.Bool("stackport.enabled", true, "Scrape stack port metrics") 
	devices            []*connector.Device
	cfg                *config.Config
=======
	stackportEnabled   = flag.Bool("stackport.enabled", true, "Scrape stack port metrics")
	etherchannelEnabled  = flag.Bool("etherchannel.enabled", true, "Scrape EtherChannel metrics")
	slottempEnabled      = flag.Bool("slottemp.enabled", true, "Scrape slot temperature metrics")
	loginfailuresEnabled = flag.Bool("loginfailures.enabled", true, "Scrape login failures metrics")
	configlogEnabled     = flag.Bool("configlog.enabled", true, "Scrape config log metrics")
	spanningtreeEnabled  = flag.Bool("spanningtree.enabled", true, "Scrape spanning tree metrics")
	poeEnabled           = flag.Bool("poe.enabled", true, "Scrape PoE metrics")
	processesEnabled     = flag.Bool("processes.enabled", true, "Scrape processes metrics")
	arpEnabled           = flag.Bool("arp.enabled", true, "Scrape ARP metrics")
	cefEnabled           = flag.Bool("cef.enabled", true, "Scrape CEF metrics")

	cachedMetrics     map[string][]prometheus.Metric
	cachedMetricsLock sync.RWMutex
	collectionInterval time.Duration = 10 * time.Second 
	cfg                *config.Config
	devices            []*connector.Device
)

const prefix = "cisco_"

var (
	upDesc                      *prometheus.Desc
	scrapeDurationDesc          *prometheus.Desc
	scrapeCollectorDurationDesc *prometheus.Desc
>>>>>>> e3bc6f2f2b94b325bd962a9f2c75adafe7e24066
)

func init() {
	upDesc = prometheus.NewDesc(prefix+"up", "Scrape of target was successful", []string{"target"}, nil)
	scrapeDurationDesc = prometheus.NewDesc(prefix+"collector_duration_seconds", "Duration of a collector scrape for one target", []string{"target"}, nil)
	scrapeCollectorDurationDesc = prometheus.NewDesc(prefix+"collect_duration_seconds", "Duration of a scrape by collector and target", []string{"target", "collector"}, nil)
}

func backgroundCollector(devices []*connector.Device, cfg *config.Config) {
	for {
		start := time.Now()
		metrics := collectMetrics(devices, cfg)
		cachedMetricsLock.Lock()
		cachedMetrics = metrics
		cachedMetricsLock.Unlock()
		duration := time.Since(start)
		if duration > collectionInterval {
			log.Warnf("Collecting metrics took %v, which is longer than the interval %v", duration, collectionInterval)
		}
		time.Sleep(collectionInterval)
	}
}

func collectMetrics(devices []*connector.Device, cfg *config.Config) map[string][]prometheus.Metric {
	metrics := make(map[string][]prometheus.Metric)
	var wg sync.WaitGroup

	for _, device := range devices {
		wg.Add(1)
		go func(d *connector.Device) {
			defer wg.Done()
			deviceMetrics := collectDeviceMetrics(d, cfg)
			cachedMetricsLock.Lock()
			metrics[d.Host] = deviceMetrics
			cachedMetricsLock.Unlock()
		}(device)
	}
	wg.Wait()
	return metrics
}

func collectDeviceMetrics(device *connector.Device, cfg *config.Config) []prometheus.Metric {
	var metrics []prometheus.Metric
	l := []string{device.Host}

	conn, err := connector.NewSSSHConnection(device, cfg)
	if err != nil {
		log.Errorln(err)
		metrics = append(metrics, prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0, l...))
		return metrics
	}
	defer conn.Close()

	metrics = append(metrics, prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1, l...))
	client := rpc.NewClient(conn, cfg.Debug)
	err = client.Identify()
	if err != nil {
		log.Errorln(device.Host + ": " + err.Error())
		return metrics
	}

	collectors := collectorsForDevices([]*connector.Device{device}, cfg)
	for _, col := range collectors.collectorsForDevice(device) {
		ct := time.Now()
		ch := make(chan prometheus.Metric)
		go func(c collector.RPCCollector) {
			err := c.Collect(client, ch, l)
			if err != nil && err.Error() != "EOF" {
				log.Errorln(c.Name() + ": " + err.Error())
			}
			close(ch)
		}(col)

		for m := range ch {
			metrics = append(metrics, m)
		}
		metrics = append(metrics, prometheus.MustNewConstMetric(scrapeCollectorDurationDesc, prometheus.GaugeValue, time.Since(ct).Seconds(), append(l, col.Name())...))
	}
	return metrics
}

func handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	cachedMetricsLock.RLock()
	defer cachedMetricsLock.RUnlock()

	reg := prometheus.NewRegistry()
	for _, deviceMetrics := range cachedMetrics {
		for _, m := range deviceMetrics {
			reg.MustRegister(&metricWrapper{Metric: m})
		}
	}

	promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog:      log.StandardLogger(),
		ErrorHandling: promhttp.ContinueOnError,
	}).ServeHTTP(w, r)
}

type metricWrapper struct {
	prometheus.Metric
}

func (m *metricWrapper) Desc() *prometheus.Desc {
	return m.Metric.Desc()
}

func (m *metricWrapper) Write(out *prometheus.Metric) error {
	return nil
}

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	err := initialize()
	if err != nil {
		log.Fatalf("Error in initialization: %v", err)
	}

	go backgroundCollector(devices, cfg)
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
	if len(*configFile) == 0 {
		log.Infoln("Loading config flags")
		return loadConfigFromFlags(), nil
	}

	log.Infoln("Loading config from", *configFile)
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
	f.StackPort = stackportEnabled
	f.BGP = bgpEnabled
	f.Environment = environmentEnabled
	f.Facts = factsEnabled
	f.Interfaces = interfacesEnabled
	f.Optics = opticsEnabled

	return c
}

func printVersion() {
	fmt.Println("cisco_exporter")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): Martin Poppen")
	fmt.Println("Metric exporter for switches and routers running cisco IOS/NX-OS/IOS-XE")
}

func startServer() {
	log.Infof("Starting Cisco exporter (Version: %s)\n", version)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Cisco Exporter (Version ` + version + `)</title></head>
			<body>
			<h1>Cisco Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			<h2>More information:</h2>
			<p><a href="https://github.com/moeinshahcheraghi/cisco_exporter">github.com/moeinshahcheraghi/cisco_exporter</a></p>
			</body>
			</html>`))
	})
	http.HandleFunc(*metricsPath, handleMetricsRequest)

	log.Infof("Listening for %s on %s\n", *metricsPath, *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}