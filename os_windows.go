package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/mackerelio/mackerel-agent/command"
	"github.com/mackerelio/mackerel-agent/mackerel"
)

func resolveConfig() (config mackerel.Config) {
	conffile := flag.String("conf", "/etc/mackerel-agent/mackerel-agent.conf", "Config file path (Configs in this file are over-written by command line options)")
	apibase := flag.String("apibase", mackerel.DefaultConfig.Apibase, "API base")
	pidfile := flag.String("pidfile", mackerel.DefaultConfig.Pidfile, "File containing PID")
	root := flag.String("root", mackerel.DefaultConfig.Root, "Directory containing variable state information")
	apikey := flag.String("apikey", "", "API key from mackerel.io web site")

	var verbose bool
	flag.BoolVar(&verbose, "verbose", mackerel.DefaultConfig.Verbose, "Toggle verbosity")
	flag.BoolVar(&verbose, "v", mackerel.DefaultConfig.Verbose, "Toggle verbosity (shorthand)")

	// The value of "role" option is internally "roll fullname",
	// but we call it "role" here for ease.
	var roleFullnames roleFullnamesFlag
	flag.Var(&roleFullnames, "role", "Set this host's roles (format: <service>:<role>)")

	flag.Parse()

	config, confErr := mackerel.LoadConfig(*conffile)
	if confErr != nil {
		logger.Criticalf("Failed to load the config file: %s", confErr)
		os.Exit(1)
	}

	// overwrite config from file by config from args
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "apibase":
			config.Apibase = *apibase
		case "apikey":
			config.Apikey = *apikey
		case "pidfile":
			config.Pidfile = *pidfile
		case "root":
			config.Root = *root
		case "verbose", "v":
			config.Verbose = verbose
		case "role":
			config.Roles = roleFullnames
		}
	})

	return
}

func start(config mackerel.Config) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		for sig := range c {
			if sig == syscall.SIGHUP {
				// nop
				// TODO reload configuration file
				logger.Debugf("Received signal '%v'", sig)
			} else {
				logger.Infof("Received signal '%v', exiting", sig)
				os.Exit(0)
			}
		}
	}()

	command.Run(config)
	return nil
}
