package spec

import (
	"github.com/mackerelio/mackerel-agent/logging"
)

type FilesystemGenerator struct {
}

func (g *FilesystemGenerator) Key() string {
	return "filesystem"
}

var filesystemLogger = logging.GetLogger("spec.filesystem")
