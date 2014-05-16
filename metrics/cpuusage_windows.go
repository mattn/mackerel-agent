package metrics

import (
	"time"
)

type CpuusageGenerator struct {
	Interval time.Duration
}

func (g *CpuusageGenerator) Generate() (Values, error) {
	interval := g.Interval * time.Second
	time.Sleep(interval)

	// TODO
	return Values{}, nil
}
