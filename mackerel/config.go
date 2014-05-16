package mackerel

import (
	"github.com/BurntSushi/toml"
	"github.com/mackerelio/mackerel-agent/logging"

	"io/ioutil"
	"os"
	"path/filepath"
)

var configLogger = logging.GetLogger("config")

const idFileName = "id"

func IdFilePath(root string) string {
	return filepath.Join(root, idFileName)
}

func LoadHostId(root string) (string, error) {
	content, err := ioutil.ReadFile(IdFilePath(root))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func SaveHostId(root string, id string) error {
	err := os.MkdirAll(root, 0744)
	if err != nil {
		return err
	}

	file, err := os.Create(IdFilePath(root))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(id))
	if err != nil {
		return err
	}

	return nil
}

type Config struct {
	Apibase         string
	Apikey          string
	Root            string
	Pidfile         string
	Roles           []string
	Verbose         bool
	Plugin          map[string]PluginConfigs
	DeprecatedSensu map[string]PluginConfigs `toml:"sensu"` // DEPRECATED this is for backward compatibility
}

type PluginConfigs map[string]PluginConfig

type PluginConfig struct {
	Command string
}

func LoadConfig(conffile string) (Config, error) {
	config, err := LoadConfigFile(conffile)

	// set default values if config does not have values
	if config.Apibase == "" {
		config.Apibase = DefaultConfig.Apibase
	}
	if config.Root == "" {
		config.Root = DefaultConfig.Root
	}
	if config.Pidfile == "" {
		config.Pidfile = DefaultConfig.Pidfile
	}
	if config.Verbose == false {
		config.Verbose = DefaultConfig.Verbose
	}

	return config, err
}

func LoadConfigFile(file string) (Config, error) {
	var config Config
	if _, err := toml.DecodeFile(file, &config); err != nil {
		return config, err
	}

	// for backward compatibility
	// merges sensu configs to plugin configs
	if _, ok := config.DeprecatedSensu["checks"]; ok {
		configLogger.Warningf("'sensu.checks.*' config format is DEPRECATED. Please use 'plugin.metrics.*' format.")

		if config.Plugin == nil {
			config.Plugin = map[string]PluginConfigs{}
		}
		if _, ok := config.Plugin["metrics"]; !ok {
			config.Plugin["metrics"] = PluginConfigs{}
		}
		for k, v := range config.DeprecatedSensu["checks"] {
			config.Plugin["metrics"]["DEPRECATED-sensu-"+k] = v
		}
	}

	return config, nil
}
