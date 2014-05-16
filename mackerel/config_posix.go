// +build !windows

package mackerel

var DefaultConfig = &Config{
	Apibase: "https://mackerel.io",
	Root:    "/var/lib/mackerel-agent",
	Pidfile: "/var/run/mackerel-agent.pid",
	Roles:   []string{},
	Verbose: false,
}
