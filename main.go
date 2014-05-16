package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mackerelio/mackerel-agent/logging"
	"github.com/mackerelio/mackerel-agent/version"
)

// allow options like -role=... -role=...
type roleFullnamesFlag []string

var roleFullnamePattern = regexp.MustCompile(`^[\w-]+:\s*[\w-]+$`)

func (r *roleFullnamesFlag) String() string {
	return fmt.Sprint(*r)
}

func (r *roleFullnamesFlag) Set(input string) error {
	inputRoles := strings.Split(input, ",")

	for _, inputRole := range inputRoles {
		if roleFullnamePattern.MatchString(inputRole) == false {
			return fmt.Errorf("Bad format for role fullname (expecting <service>:<role>): %s", inputRole)
		}
	}

	*r = append(*r, inputRoles...)

	return nil
}

var logger = logging.GetLogger("main")

func main() {
	config := resolveConfig()

	if config.Verbose {
		logging.ConfigureLoggers("DEBUG")
	} else {
		logging.ConfigureLoggers("INFO")
	}

	logger.Infof("Starting mackerel-agent version:%s, rev:%s", version.VERSION, version.GITCOMMIT)

	if config.Apikey == "" {
		logger.Criticalf("Apikey must be specified in the command-line flag or in the config file")
		os.Exit(1)
	}

	if err := start(config); err != nil {
		os.Exit(1)
	}
}
