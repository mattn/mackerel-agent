package metrics

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
)

var memItems = map[string]*regexp.Regexp{
	"total":    regexp.MustCompile(`^MemTotal:\s+(\d+) (.+)$`),
	"free":     regexp.MustCompile(`^MemFree:\s+(\d+) (.+)$`),
	"buffers":  regexp.MustCompile(`^Buffers:\s+(\d+) (.+)$`),
	"cached":   regexp.MustCompile(`^Cached:\s+(\d+) (.+)$`),
	"active":   regexp.MustCompile(`^Active:\s+(\d+) (.+)$`),
	"inactive": regexp.MustCompile(`^Inactive:\s+(\d+) (.+)$`),
	// "high_total":       regexp.MustCompile(`^HighTotal:\s+(\d+) (.+)$`),
	// "high_free":        regexp.MustCompile(`^HighFree:\s+(\d+) (.+)$`),
	// "low_total":        regexp.MustCompile(`^LowTotal:\s+(\d+) (.+)$`),
	// "low_free":         regexp.MustCompile(`^LowFree:\s+(\d+) (.+)$`),
	// "dirty":            regexp.MustCompile(`^Dirty:\s+(\d+) (.+)$`),
	// "writeback":        regexp.MustCompile(`^Writeback:\s+(\d+) (.+)$`),
	// "anon_pages":       regexp.MustCompile(`^AnonPages:\s+(\d+) (.+)$`),
	// "mapped":           regexp.MustCompile(`^Mapped:\s+(\d+) (.+)$`),
	// "slab":             regexp.MustCompile(`^Slab:\s+(\d+) (.+)$`),
	// "slab_reclaimable": regexp.MustCompile(`^SReclaimable:\s+(\d+) (.+)$`),
	// "slab_unreclaim":   regexp.MustCompile(`^SUnreclaim:\s+(\d+) (.+)$`),
	// "page_tables":      regexp.MustCompile(`^PageTables:\s+(\d+) (.+)$`),
	// "nfs_unstable":     regexp.MustCompile(`^NFS_Unstable:\s+(\d+) (.+)$`),
	// "bounce":           regexp.MustCompile(`^Bounce:\s+(\d+) (.+)$`),
	// "commit_limit":     regexp.MustCompile(`^CommitLimit:\s+(\d+) (.+)$`),
	// "committed_as":     regexp.MustCompile(`^Committed_AS:\s+(\d+) (.+)$`),
	// "vmalloc_total":    regexp.MustCompile(`^VmallocTotal:\s+(\d+) (.+)$`),
	// "vmalloc_used":     regexp.MustCompile(`^VmallocUsed:\s+(\d+) (.+)$`),
	// "vmalloc_chunk":    regexp.MustCompile(`^VmallocChunk:\s+(\d+) (.+)$`),
	"swap_cached": regexp.MustCompile(`^SwapCached:\s+(\d+) (.+)$`),
	"swap_total":  regexp.MustCompile(`^SwapTotal:\s+(\d+) (.+)$`),
	"swap_free":   regexp.MustCompile(`^SwapFree:\s+(\d+) (.+)$`),
}

/*
collect memory usage


`memory.{metric}`: using memory size[KiB] retrieved from /proc/meminfo

metric = "total", "free", "buffers", "cached", "active", "inactive", "swap_cached", "swap_total", "swap_free"

Metrics "used" is caluculated here like (total - free - buffers - cached) for ease.
This calculation may be going to be done in server side in the future.

graph: stacks `memory.{metric}`
*/
type MemoryGenerator struct {
}

func (g *MemoryGenerator) Generate() (Values, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		memoryLogger.Errorf("Failed (skip these metrics): %s", err)
		return nil, err
	}
	scanner := bufio.NewScanner(file)

	ret := make(map[string]float64)
	used := float64(0)
	used_cnt := 0
	for scanner.Scan() {
		line := scanner.Text()
		for k, regexp := range memItems {
			if matches := regexp.FindStringSubmatch(line); matches != nil {
				// ex.) MemTotal:        3916792 kB
				// matches[1] = 3916792, matches[2] = kB
				if matches[2] != "kB" {
					memoryLogger.Warningf("/proc/meminfo contains an invalid unit: %s", k)
					break
				}
				value, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					memoryLogger.Warningf("Failed to parse memory metrics: %s", err)
					break
				}
				ret["memory."+k] = value * 1024
				if k == "free" || k == "buffers" || k == "cached" {
					used -= value
					used_cnt++
				}
				if k == "total" {
					used += value
					used_cnt++
				}
				break
			}
		}
	}
	if err := scanner.Err(); err != nil {
		memoryLogger.Errorf("Failed (skip these metrics): %s", err)
		return nil, err
	}
	if used_cnt == 4 { // 4 is free, buffers, cached and total
		ret["memory.used"] = used * 1024
	}

	return Values(ret), nil
}
