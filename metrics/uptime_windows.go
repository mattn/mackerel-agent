package metrics

import (
	. "github.com/mackerelio/mackerel-agent/util"
)

/*
collect uptime

`uptime`: uptime[day] retrieved from /proc/uptime

graph: `uptime`
*/
type UptimeGenerator struct {
}

func (g *UptimeGenerator) Generate() (Values, error) {
	r, _, err := GetTickCount.Call()
	if r == 0 {
		return nil, err
	}

	return Values(map[string]float64{"uptime": float64(r) / 1000}), nil
}
