func (c *opticsCollector) ParseAllTransceivers(ostype string, output string) (map[string]Optics, error) {
	items := make(map[string]Optics)

	if ostype == rpc.NXOS {
		// همان قبلی (بدون تغییر)
		sections := strings.Split(output, "\n\n")
		reTx := regexp.MustCompile(`Tx Power\s*((?:-)?[\d\.]+)`)
		reRx := regexp.MustCompile(`Rx Power\s*((?:-)?[\d\.]+)`)
		for _, section := range sections {
			lines := strings.Split(section, "\n")
			if len(lines) > 0 && strings.HasPrefix(lines[0], "Ethernet") {
				iface := strings.TrimSpace(lines[0])
				var tx, rx float64
				for _, line := range lines {
					if m := reTx.FindStringSubmatch(line); m != nil {
						tx = util.Str2float64(m[1])
					} else if m := reRx.FindStringSubmatch(line); m != nil {
						rx = util.Str2float64(m[1])
					}
				}
				items[iface] = Optics{RxPower: rx, TxPower: tx}
			}
		}

	} else if ostype == rpc.IOS || ostype == rpc.IOSXE {
		// پارس یکسان برای IOS و IOS-XE
		lines := strings.Split(output, "\n")
		// regex خیلی انعطاف‌پذیر (پشتیبانی از N/A, --, -inf و غیره)
		re := regexp.MustCompile(`^\s*(\S[\S\s]*?\S)\s+([-\d.infNA]+)\s+([-\d.infNA]+)`)
		for _, line := range lines {
			if strings.Contains(line, "Port") || strings.Contains(line, "Interface") || strings.Contains(line, "---") || strings.TrimSpace(line) == "" {
				continue
			}
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 4 {
				tx := util.Str2float64(matches[2])
				rx := util.Str2float64(matches[3])
				// اگر مقدار نامعتبر بود (مثل N/A) → -inf می‌شود → ما -40 می‌گذاریم (یا می‌تونی 0 بگذاری)
				if tx < -40 {
					tx = -40
				}
				if rx < -40 {
					rx = -40
				}
				items[strings.TrimSpace(matches[1])] = Optics{TxPower: tx, RxPower: rx}
			}
		}
	}

	return items, nil
}