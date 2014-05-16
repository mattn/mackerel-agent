package metrics

import (
	"syscall"
	"unsafe"

	. "github.com/mackerelio/mackerel-agent/util"
)

type Loadavg5Generator struct {
	query   syscall.Handle
	counters []*CounterInfo
}

func (g *Loadavg5Generator) Generate() (Values, error) {
	if g.query == 0 {
		var err error
		g.query, err = CreateQuery()
		if err != nil {
			return nil, err
		}

		counter, err := CreateCounter(g.query, "loadavg5", `\Processor(_Total)\% Processor Time`)
		if err != nil {
			return nil, err
		}
		g.counters = append(g.counters, counter)
	}

	r, _, err := PdhCollectQueryData.Call(uintptr(g.query))
	if r != 0 {
		return nil, err
	}

	results := make(map[string]float64)
	for _, v := range g.counters {
		var value PDH_FMT_COUNTERVALUE_ITEM_DOUBLE
		r, _, err = PdhGetFormattedCounterValue.Call(uintptr(v.Counter), PDH_FMT_DOUBLE, uintptr(0), uintptr(unsafe.Pointer(&value)))
		if r != 0 && r != PDH_INVALID_DATA {
			return nil, err
		}
		results[v.PostName] = value.FmtValue.DoubleValue
	}
	return results, nil
}
