package metrics

import (
	"unsafe"

	. "github.com/mackerelio/mackerel-agent/util"
)

type MemoryGenerator struct {
}

func (g *MemoryGenerator) Generate() (Values, error) {
	ret := make(map[string]float64)

	var memoryStatusEx MEMORYSTATUSEX
	memoryStatusEx.Length = uint32(unsafe.Sizeof(memoryStatusEx))
	r, _, err := GlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memoryStatusEx)))
	if r == 0 {
		return nil, err
	}

	ret["memory.total"] = float64(memoryStatusEx.TotalPhys) / 1024
	ret["memory.free"] = float64(memoryStatusEx.AvailPhys) / 1024

	return Values(ret), nil
}
