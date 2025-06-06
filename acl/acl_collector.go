package acl

import (
    "regexp"
    "strings"
    "strconv"
    "github.com/moeinshahcheraghi/cisco_exporter/collector"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
    "github.com/prometheus/client_golang/prometheus"
)

const prefix = "cisco_acl_"

var (
    hitsDesc *prometheus.Desc
)

func init() {
    hitsDesc = prometheus.NewDesc(prefix+"hits", "Number of hits per ACL rule", []string{"target", "acl", "rule"}, nil)
}

type aclCollector struct{}

func NewCollector() collector.RPCCollector {
    return &aclCollector{}
}

func (*aclCollector) Name() string {
    return "ACL"
}

func (*aclCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- hitsDesc
}

func (c *aclCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
    out, err := client.RunCommand("show access-lists")
    if err != nil {
        return err
    }
    hits, err := parseACLHits(out)
    if err != nil {
        return err
    }
    for _, hit := range hits {
        labels := append(labelValues, hit.ACL, hit.Rule)
        ch <- prometheus.MustNewConstMetric(hitsDesc, prometheus.CounterValue, float64(hit.Count), labels...)
    }
    return nil
}

type Hit struct {
    ACL   string
    Rule  string
    Count int
}

func parseACLHits(output string) ([]Hit, error) {
    var hits []Hit
    lines := strings.Split(output, "\n")
    re := regexp.MustCompile(`access-list (\S+)\s+.*\s+rule (\d+)\s+.*\((\d+) matches\)`)
    for _, line := range lines {
        matches := re.FindStringSubmatch(line)
        if matches != nil {
            count, _ := strconv.Atoi(matches[3])
            hits = append(hits, Hit{
                ACL:   matches[1],
                Rule:  matches[2],
                Count: count,
            })
        }
    }
    return hits, nil
}
