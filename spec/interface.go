package spec

import (
	"github.com/mackerelio/mackerel-agent/logging"
)

type InterfaceGenerator struct {
}

func (g *InterfaceGenerator) Key() string {
	return "interface"
}

var interfaceLogger = logging.GetLogger("spec.interface")
