package metrics

import (
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"
	"unsafe"

	. "github.com/mackerelio/mackerel-agent/util"
)

type InterfaceGenerator struct {
	Interval time.Duration
	query    syscall.Handle
	counters []*CounterInfo
}

func (g *InterfaceGenerator) Generate() (Values, error) {
	if g.query == 0 {
		var err error
		g.query, err = CreateQuery()
		if err != nil {
			return nil, err
		}

		ifs, err := net.Interfaces()
		if err != nil {
			return nil, err
		}

		ai, err := GetAdapterList()
		if err != nil {
			return nil, err
		}

		for _, ifi := range ifs {
			for ; ai != nil; ai = ai.Next {
				if ifi.Index == int(ai.Index) {
					name := BytePtrToString(&ai.Description[0])
					name = strings.Replace(name, "(", "[", -1)
					name = strings.Replace(name, ")", "]", -1)
					var counter *CounterInfo

					counter, err = CreateCounter(
						g.query,
						fmt.Sprintf(`interface.nic%d.rxBytes.delta`, ifi.Index),
						fmt.Sprintf(`\Network Interface(%s)\Bytes Received/sec`, name))
					if err != nil {
						return nil, err
					}
					g.counters = append(g.counters, counter)
					counter, err = CreateCounter(
						g.query,
						fmt.Sprintf(`interface.nic%d.txBytes.delta`, ifi.Index),
						fmt.Sprintf(`\Network Interface(%s)\Bytes Sent/sec`, name))
					if err != nil {
						return nil, err
					}
					g.counters = append(g.counters, counter)
				}
			}
		}

		r, _, err := PdhCollectQueryData.Call(uintptr(g.query))
		if r != 0 {
			return nil, err
		}
	}

	interval := g.Interval * time.Second
	time.Sleep(interval)

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
