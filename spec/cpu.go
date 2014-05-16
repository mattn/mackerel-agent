package spec

import (
	"github.com/mackerelio/mackerel-agent/logging"
)

var cpuLogger = logging.GetLogger("spec.cpu")

type CPUGenerator struct {
}

func (g CPUGenerator) Key() string {
	return "cpu"
}
