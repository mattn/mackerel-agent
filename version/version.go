package version

import (
	"fmt"
)

// make build sets this automaticaly
var VERSION string

// make build sets this automaticaly
var GITCOMMIT string

func UserAgent() string {
	return fmt.Sprintf("mackerel-agent/%s (Revision %s)", VERSION, GITCOMMIT)
}
