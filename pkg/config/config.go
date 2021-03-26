package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// SaveConfig writes yaml to a file
func (c *Config) SaveConfig(filename string) error {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, bytes, 0644)
}

func (c *Config) String() string {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}
	return string(bytes)
}
