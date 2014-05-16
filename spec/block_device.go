package spec

import (
	"github.com/mackerelio/mackerel-agent/logging"
)

type BlockDeviceGenerator struct {
}

func (g *BlockDeviceGenerator) Key() string {
	return "block_device"
}

var blockDeviceLogger = logging.GetLogger("spec.block_device")
