package metrics

import (
	"errors"
	"time"
)

type CpuusageGenerator struct {
	Interval time.Duration
}

func (g *CpuusageGenerator) Generate() (Values, error) {
	return nil, errors.New("Not Implemented")
}
