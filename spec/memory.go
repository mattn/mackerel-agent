package spec

import (
	"github.com/mackerelio/mackerel-agent/logging"
)

type MemoryGenerator struct {
}

func (g *MemoryGenerator) Key() string {
	return "memory"
}

var memoryLogger = logging.GetLogger("spec.memory")
