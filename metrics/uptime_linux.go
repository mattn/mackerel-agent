package metrics

import (
	"io/ioutil"
	"strconv"
	"strings"
)

/*
collect uptime

`uptime`: uptime[day] retrieved from /proc/uptime

graph: `uptime`
*/
type UptimeGenerator struct {
}

func (g *UptimeGenerator) Generate() (Values, error) {
	contentbytes, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		uptimeLogger.Errorf("Failed (skip these metrics): %s", err)
		return nil, err
	}
	content := string(contentbytes)
	cols := strings.Split(content, " ")

	f, err := strconv.ParseFloat(cols[0], 64)
	if err != nil {
		uptimeLogger.Errorf("Failed to parse values (skip these metrics): %s", err)
		return nil, err
	}

	return Values(map[string]float64{"uptime": f / 86400}), nil
}
