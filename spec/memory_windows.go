package spec

import (
	"fmt"
	"unsafe"

	. "github.com/mackerelio/mackerel-agent/util"
)

func (g *MemoryGenerator) Generate() (interface{}, error) {
	result := make(map[string]interface{})

	var memoryStatusEx MEMORYSTATUSEX
	memoryStatusEx.Length = uint32(unsafe.Sizeof(memoryStatusEx))
	r, _, err := GlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memoryStatusEx)))
	if r == 0 {
		return nil, err
	}

	result["total"] = fmt.Sprintf("%dkb", memoryStatusEx.TotalPhys/1024)
	result["free"] = fmt.Sprintf("%dkb", memoryStatusEx.AvailPhys/1024)

	return result, nil
}
