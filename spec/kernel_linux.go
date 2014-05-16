package spec

import (
	"os/exec"
	"strings"
)

func (g *KernelGenerator) Generate() (interface{}, error) {
	commands := map[string][]string{
		"name":    []string{"uname", "-s"},
		"release": []string{"uname", "-r"},
		"version": []string{"uname", "-v"},
		"machine": []string{"uname", "-m"},
		"os":      []string{"uname", "-o"},
	}

	results := make(map[string]string)
	for key, command := range commands {
		out, err := exec.Command(command[0], command[1]).Output()
		if err != nil {
			kernelLogger.Errorf("Failed to run %s %s (skip this spec): %s", command[0], command[1], err)
			return nil, err
		}
		str := strings.TrimSpace(string(out))

		results[key] = str
	}

	return results, nil
}

