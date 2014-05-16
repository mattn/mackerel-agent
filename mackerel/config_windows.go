package mackerel

var DefaultConfig = &Config{
	Apibase: "https://mackerel.io",
	Root:    ".",
	Pidfile: "pid",
	Roles:   []string{},
	Verbose: false,
}
