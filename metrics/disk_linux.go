package metrics

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mackerelio/mackerel-agent/logging"
)

/*
collect disk I/O

`disk.{device}.{metric}.delta`: The increased amount of disk I/O per minute retrieved from /proc/diskstats

`disk.{device}.{metric}`: [DEPRECATED] The total amount of disk I/O retrieved from /proc/diskstats

device = "sda1", "xvda1" and so on...

metric = "reads", "readsMerged", "sectorsRead", "readTime", "writes", "writesMerged", "sectorsWritten", "writeTime", "ioInProgress", "ioTime", "ioTimeWeighted"

graph: `disk.{device}.{metric}.delta`

cat /proc/diskstats sample:
	202       1 xvda1 750193 3037 28116978 368712 16600606 7233846 424712632 23987908 0 2355636 24345740
	202       2 xvda2 1641 9310 87552 1252 6365 3717 80664 24192 0 15040 25428
	  7       0 loop0 0 0 0 0 0 0 0 0 0 0 0
	  7       1 loop1 0 0 0 0 0 0 0 0 0 0 0
	253       0 dm-0 46095806 0 549095028 2243928 7192424 0 305024576 12521088 0 2728444 14782668
	253     628 dm-628 3198 0 75410 1360 30802835 0 3942653176 1334317408 0 70948 1358596768
253       2 dm-2 2022 0 42250 488 30822403 0 3942809696 1364721232 0 93348 1382989868
*/
type DiskGenerator struct {
	Interval time.Duration
}

var diskMetricsNames = []string{
	"reads", "readsMerged", "sectorsRead", "readTime",
	"writes", "writesMerged", "sectorsWritten", "writeTime",
	"ioInProgress", "ioTime", "ioTimeWeighted",
}

func (g *DiskGenerator) Generate() (Values, error) {
	prevValues, err := g.collectDiskstatValues()
	if err != nil {
		return nil, err
	}

	interval := g.Interval * time.Second
	time.Sleep(interval)

	currValues, err := g.collectDiskstatValues()
	if err != nil {
		return nil, err
	}

	for name, value := range prevValues {
		currValue, ok := currValues[name]
		if ok {
			currValues[name+".delta"] = (currValue - value) / interval.Seconds()
		}
	}

	return currValues, nil
}

func (g *DiskGenerator) collectDiskstatValues() (Values, error) {
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		diskLogger.Errorf("Failed (skip these metrics): %s", err)
		return nil, err
	}

	lineScanner := bufio.NewScanner(bufio.NewReader(file))
	results := make(map[string]float64)
	for lineScanner.Scan() {
		cols := strings.Fields(lineScanner.Text())
		device := cols[2]
		values := cols[3:]

		deviceResult := make(map[string]float64)
		hasNonZeroValue := false
		for i, _ := range diskMetricsNames {
			key := fmt.Sprintf("disk.%s.%s", device, diskMetricsNames[i])
			value, err := strconv.ParseFloat(values[i], 64)
			if err != nil {
				diskLogger.Warningf("Failed to parse disk metrics: %s", err)
				break
			}
			if value != 0 {
				hasNonZeroValue = true
			}
			deviceResult[key] = value
		}
		if hasNonZeroValue {
			for k, v := range deviceResult {
				results[k] = v
			}
		}
	}

	return results, nil
}
