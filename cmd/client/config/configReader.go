package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

const (
	defaultPath = "config.json"
)

func ReadConfig(configFilePath string) (*Config, error) {
	if len(strings.TrimSpace(configFilePath)) == 0 {
		configFilePath = defaultPath
	}
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}

	defer configFile.Close()

	jsonBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	var config Config = DefaultConfig
	err = json.Unmarshal(jsonBytes, &config)
	return &config, err
}
