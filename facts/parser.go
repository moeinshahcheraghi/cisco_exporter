package facts

import (
	"errors"
	"regexp"
	"strings"

	"github.com/moeinshahcheraghi/cisco_exporter/rpc"
	"github.com/moeinshahcheraghi/cisco_exporter/util"
)

// ParseVersion parses cli output and tries to find the version number of the running OS
func (c *factsCollector) ParseVersion(ostype string, output string) (VersionFact, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return VersionFact{}, errors.New("'show version' is not implemented for " + ostype)
	}
	versionRegexp := make(map[string]*regexp.Regexp)
	versionRegexp[rpc.IOSXE], _ = regexp.Compile(`^.*, Version (.+) -.*$`)
	versionRegexp[rpc.IOS], _ = regexp.Compile(`^.*, Version (.+),.*$`)
	versionRegexp[rpc.NXOS], _ = regexp.Compile(`^\s+NXOS: version (.*)$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := versionRegexp[ostype].FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		return VersionFact{Version: ostype + "-" + matches[1]}, nil
	}
	return VersionFact{}, errors.New("Version string not found")
}

// ParseMemory parses cli output and tries to find current memory usage (IOS / IOS-XE)
func (c *factsCollector) ParseMemory(ostype string, output string) ([]MemoryFact, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return nil, errors.New("'show process memory' is not implemented for " + ostype)
	}
	memoryRegexp, _ := regexp.Compile(`^\s*(\S*) Pool Total:\s*(\d+) Used:\s*(\d+) Free:\s*(\d+)\s*$`)

	items := []MemoryFact{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := memoryRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		item := MemoryFact{
			Type:  matches[1],
			Total: util.Str2float64(matches[2]),
			Used:  util.Str2float64(matches[3]),
			Free:  util.Str2float64(matches[4]),
		}
		items = append(items, item)
	}
	return items, nil
}

// ParseCPU parses cli output and tries to find current CPU utilization (IOS / IOS-XE)
func (c *factsCollector) ParseCPU(ostype string, output string) (CPUFact, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return CPUFact{}, errors.New("'show process cpu' is not implemented for " + ostype)
	}
	memoryRegexp, _ := regexp.Compile(`^\s*CPU utilization for five seconds: (\d+)%\/(\d+)%; one minute: (\d+)%; five minutes: (\d+)%.*$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := memoryRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		return CPUFact{
			FiveSeconds: util.Str2float64(matches[1]),
			Interrupts:  util.Str2float64(matches[2]),
			OneMinute:   util.Str2float64(matches[3]),
			FiveMinutes: util.Str2float64(matches[4]),
		}, nil
	}
	return CPUFact{}, errors.New("Version string not found")
}

// ParseSystemResources parses NX-OS 'show system resources' output and
// returns both CPU and Memory facts from a single command.
//
// Typical output:
//
//	Load average:   1 minute: 0.15   5 minutes: 0.10   15 minutes: 0.08
//	Processes   :   105 total, 1 running
//	CPU states  :   2.5% user,   1.0% kernel,   96.5% idle
//	Memory usage:   16401192K total,  9845284K used,   6555908K free
//	Current memory status: OK
func (c *factsCollector) ParseSystemResources(output string) (CPUFact, []MemoryFact, error) {
	cpuRegexp := regexp.MustCompile(`(?i)CPU states\s*:\s*([\d.]+)%\s*user,\s*([\d.]+)%\s*kernel,\s*([\d.]+)%\s*idle`)
	memRegexp := regexp.MustCompile(`(?i)Memory usage\s*:\s*(\d+)K\s*total,\s*(\d+)K\s*used,\s*(\d+)K\s*free`)

	cpuMatches := cpuRegexp.FindStringSubmatch(output)
	if cpuMatches == nil {
		return CPUFact{}, nil, errors.New("CPU utilization not found in 'show system resources' output")
	}

	kernel := util.Str2float64(cpuMatches[2])
	idle := util.Str2float64(cpuMatches[3])
	used := 100 - idle
	if used < 0 {
		used = 0
	}

	cpu := CPUFact{
		FiveSeconds: used,
		OneMinute:   used,
		FiveMinutes: used,
		Interrupts:  kernel,
	}

	items := []MemoryFact{}
	if memMatches := memRegexp.FindStringSubmatch(output); memMatches != nil {
		// values are reported in Kilobytes, convert to bytes to stay
		// consistent with the IOS/IOS-XE collector
		totalK := util.Str2float64(memMatches[1])
		usedK := util.Str2float64(memMatches[2])
		freeK := util.Str2float64(memMatches[3])
		items = append(items, MemoryFact{
			Type:  "System",
			Total: totalK * 1024,
			Used:  usedK * 1024,
			Free:  freeK * 1024,
		})
	}

	return cpu, items, nil
}