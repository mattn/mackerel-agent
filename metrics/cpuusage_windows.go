package metrics

import (
	"time"
)

type CpuusageGenerator struct {
	Interval time.Duration
}

var cpuusageMetricNames = []string{}

func (g *CpuusageGenerator) Generate() (Values, error) {
	interval := g.Interval * time.Second
	time.Sleep(interval)

	// TODO
	return Values{}, nil
}
