package spec

import (
	"github.com/mackerelio/mackerel-agent/logging"
)

type KernelGenerator struct {
}

func (g *KernelGenerator) Key() string {
	return "kernel"
}

var kernelLogger = logging.GetLogger("spec.kernel")
