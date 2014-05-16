package command

import (
	"fmt"
	"os"

	"github.com/mackerelio/mackerel-agent/mackerel"
	"github.com/mackerelio/mackerel-agent/spec"
)

func prepareHost(root string, api *mackerel.API, specGenerators []spec.Generator, roleFullnames []string) (*mackerel.Host, error) {
	specs := collectSpecs(specGenerators)

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain hostname: %s", err.Error())
	}

	var result *mackerel.Host
	if hostId, err := mackerel.LoadHostId(root); err != nil { // create
		logger.Debugf("Registering new host on mackerel...")
		interfaces := collectInterfaces()
		createdHostId, err := api.CreateHost(hostname, specs, interfaces, roleFullnames)
		if err != nil {
			return nil, fmt.Errorf("Failed to register this host: %s", err.Error())
		}

		result, err = api.FindHost(createdHostId)
		if err != nil {
			return nil, fmt.Errorf("Failed to find this host on mackerel: %s", err.Error())
		}
	} else { // update
		result, err = api.FindHost(hostId)
		if err != nil {
			return nil, fmt.Errorf("Failed to find this host on mackerel (You may want to delete file \"%s\" to register this host to an another organization): %s", mackerel.IdFilePath(root), err.Error())
		}
		interfaces := collectInterfaces()
		err := api.UpdateHost(hostId, hostname, specs, interfaces, roleFullnames)
		if err != nil {
			return nil, fmt.Errorf("Failed to update this host: %s", err.Error())
		}
	}

	err = mackerel.SaveHostId(root, result.Id)
	if err != nil {
		logger.Criticalf("Failed to save host ID: %s", err.Error())
		os.Exit(1)
	}

	return result, nil
}
