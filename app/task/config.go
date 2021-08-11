package task

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config defiles list of tasks
type Config struct {
	Tasks []struct {
		Name    string `yaml:"name"`
		Command string `yaml:"command"`
	} `yaml:"tasks"`
}

// LoadConfig reads and parses yaml config
func LoadConfig(file string) (*Config, error) {
	fh, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("can't load config file %s: %w", file, err)
	}
	defer fh.Close()

	res := Config{}
	if err := yaml.NewDecoder(fh).Decode(&res); err != nil {
		return nil, fmt.Errorf("can't parse config: %w", err)
	}
	return &res, nil
}

// GetTaskCommand retrieves the command for given task name
func (c *Config) GetTaskCommand(name string) (command string, ok bool) {
	for _, t := range c.Tasks {
		if strings.EqualFold(name, t.Name) {
			return t.Command, true
		}
	}
	return "", false
}
