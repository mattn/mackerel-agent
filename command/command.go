package command

import (
	"crypto/sha1"
	"os"
	"time"

	"github.com/mackerelio/mackerel-agent/agent"
	"github.com/mackerelio/mackerel-agent/logging"
	"github.com/mackerelio/mackerel-agent/mackerel"
	"github.com/mackerelio/mackerel-agent/metrics"
	"github.com/mackerelio/mackerel-agent/spec"
	"github.com/mackerelio/mackerel-agent/version"
)

var logger = logging.GetLogger("command")

func collectSpecs(specGenerators []spec.Generator) map[string]interface{} {
	specs := make(map[string]interface{})
	for _, g := range specGenerators {
		value, err := g.Generate()
		if err != nil {
			logger.Errorf("Failed to collect specs in %T (skip this spec): %s", g, err.Error())
		}
		specs[g.Key()] = value
	}
	specs["agent-version"] = version.VERSION
	specs["agent-revision"] = version.GITCOMMIT
	specs["agent-name"] = version.UserAgent()
	return specs
}

func collectInterfaces() []map[string]interface{} {
	g := &spec.InterfaceGenerator{}
	value, err := g.Generate()
	if err != nil {
		logger.Errorf("Failed to collect interfaces in %T (skip the interfaces): %s", g, err.Error())
		return nil
	}
	return value.([]map[string]interface{})
}

const METRICS_POST_DEQUEUE_DELAY = 30 * time.Second // delay for dequeuing from buffer queue
const METRICS_POST_RETRY_DELAY = 1 * time.Minute    // delay for retring a request that causes errors
const METRICS_POST_RETRY_MAX = 10                   // max numbers of retries for a request that causes errors
const METRICS_POST_BUFFER_SIZE = 30                 // max numbers of requests stored in buffer queue.

func delayByHost(host *mackerel.Host) time.Duration {
	s := sha1.Sum([]byte(host.Id))
	return time.Duration(int(s[len(s)-1])%60) * time.Second
}

func loop(ag *agent.Agent, api *mackerel.API, host *mackerel.Host) {
	metricsResult := ag.Watch()

	postQueue := make(chan []*mackerel.CreatingMetricsValue, METRICS_POST_BUFFER_SIZE)

	go func() {
		for values := range postQueue {
			tries := METRICS_POST_RETRY_MAX
			for {
				err := api.PostMetricsValues(values)
				if err == nil {
					logger.Debugf("Posting metrics succeeded.")
					break
				}
				logger.Errorf("Failed to post metrics value (will retry): %s", err.Error())

				tries -= 1
				if tries <= 0 {
					logger.Errorf("Give up retrying to post metrics.")
					break
				}

				logger.Debugf("Retrying to post metrics...")
				time.Sleep(METRICS_POST_RETRY_DELAY)
			}

			time.Sleep(METRICS_POST_DEQUEUE_DELAY)
		}
	}()

	postDelay := delayByHost(host)
	isFirstTime := true
	for {
		select {
		case result := <-metricsResult:
			created := float64(result.Created.Unix())
			creatingValues := [](*mackerel.CreatingMetricsValue){}
			for name, value := range (map[string]float64)(result.Values) {
				creatingValues = append(
					creatingValues,
					&mackerel.CreatingMetricsValue{host.Id, name, created, value},
				)
			}
			if isFirstTime { // request immediately to create graph defs of host
				isFirstTime = false
			} else {
				// Sending data at every 0 second from all hosts causes request flooding.
				// To prevent flooding, this loop sleeps for some seconds
				// which is specific to the ID of the host running agent on.
				// The sleep second is up to 60s.
				logger.Debugf("Sleeping %v to enqueue post request...", postDelay)
				time.Sleep(postDelay)
			}
			logger.Debugf("Enqueuing task to post metrics.")
			postQueue <- creatingValues
		}
	}
}

func Run(config mackerel.Config) {
	api, err := mackerel.NewApi(config.Apibase, config.Apikey, config.Verbose)
	if err != nil {
		logger.Criticalf("Failed to prepare an api: %s", err.Error())
		os.Exit(1)
	}

	specGenerators := []spec.Generator{
		&spec.KernelGenerator{},
		&spec.CPUGenerator{},
		&spec.MemoryGenerator{},
		&spec.BlockDeviceGenerator{},
		&spec.FilesystemGenerator{},
	}

	host, err := prepareHost(config.Root, api, specGenerators, config.Roles)
	if err != nil {
		logger.Criticalf("Failed to run this agent: %s", err.Error())
		os.Exit(1)
	}

	logger.Infof("Start: apibase = %s, hostName = %s, hostId = %s", config.Apibase, host.Name, host.Id)

	generators := []metrics.Generator{
		&metrics.Loadavg5Generator{},
		&metrics.CpuusageGenerator{Interval: 60},
		&metrics.MemoryGenerator{},
		&metrics.UptimeGenerator{},
		&metrics.InterfaceGenerator{Interval: 60},
		&metrics.DiskGenerator{Interval: 60},
	}

	for _, pluginConfig := range config.Plugin["metrics"] {
		generators = append(generators, &metrics.PluginGenerator{pluginConfig})
	}

	ag := &agent.Agent{generators}
	loop(ag, api, host)
}
