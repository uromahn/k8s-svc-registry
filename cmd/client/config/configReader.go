package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
)

const (
	defaultPath = "config.json"
)

func ReadConfig(configFilePath string) (*Config, error) {
	if len(strings.TrimSpace(configFilePath)) == 0 {
		configFilePath = defaultPath
	}
	klog.Infof("Opening config file %s", configFilePath)
	configFile, err := os.Open(configFilePath)
	if err != nil {
		klog.Errorf("Error while opening config file %s: %s", configFilePath, err.Error())
		return nil, err
	}

	defer configFile.Close()

	jsonBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		klog.Errorf("Error reading file %s: %s", configFilePath, err.Error())
		return nil, err
	}
	klog.Infof("Config file read with %d bytes", len(jsonBytes))

	var config Config = DefaultConfig
	err = json.Unmarshal(jsonBytes, &config)
	if err != nil {
		klog.Errorf("Error while parsing JSON config from file: %s", err.Error())
		return nil, err
	}

	// next we check if we have a serviceConfDir defined and if yes, read and parse all files from there
	if config.ServiceConfDir != nil {
		klog.Infof("scanning through ServiceConfDir '%s'", *config.ServiceConfDir)
		err = parseAndAddServicesFromFile(&config)
		if err != nil {
			klog.Errorf("Error happened while trying to parse service config files in %s: %s", *config.ServiceConfDir, err.Error())
			return nil, err
		}
	} else {
		klog.Info("No ServiceConfDir specified")
	}
	return &config, nil
}

func collectServiceConfigFiles(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			klog.Errorf("Error while walking configFile folder: %s", err.Error())
			return err
		}
		if info.IsDir() {
			// we don't want to read files recursively
			return nil
		}
		*files = append(*files, path)
		return nil
	}
}

func parseAndAddServicesFromFile(config *Config) error {
	var serviceConfigFiles []string

	path := config.ServiceConfDir
	err := filepath.Walk(*path, collectServiceConfigFiles(&serviceConfigFiles))
	if err != nil {
		klog.Errorf("Error while getting all filenames for service configs: %s", err.Error())
		return err
	}
	// now we need to look over all the files we found, read their content and parse it into JSON
	for _, serviceFilePath := range serviceConfigFiles {
		service, err := readAndParseServiceFromFile(serviceFilePath)
		if err != nil {
			// use the pure filename as service name
			serviceFileName := filepath.Base(serviceFilePath)
			serviceName := strings.TrimSuffix(serviceFileName, filepath.Ext(serviceFileName))
			if config.Services == nil {
				var services ServicesType = make(map[string]Service)
				config.Services = &services
			}
			services := *config.Services
			services[serviceName] = *service
			config.Services = &services
		} else {
			klog.Errorf("Invalid service config found in %s - skipping file: %s", serviceFilePath, err.Error())
		}
	}
	return nil
}

func readAndParseServiceFromFile(serviceFilePath string) (*Service, error) {
	if len(strings.TrimSpace(serviceFilePath)) == 0 {
		err := fmt.Errorf("got filename with zero length")
		klog.Error(err.Error())
		return nil, err
	}
	serviceFile, err := os.Open(serviceFilePath)

	if err != nil {
		klog.Errorf("Error while opening service file %s: %s", serviceFilePath, err.Error())
		return nil, err
	}

	defer serviceFile.Close()
	jsonBytes, err := ioutil.ReadAll(serviceFile)
	if err != nil {
		klog.Errorf("Error reading file %s: %s", serviceFilePath, err.Error())
		return nil, err
	}
	klog.Infof("Config file read with %d bytes", len(jsonBytes))

	var service Service = EmptyService
	err = json.Unmarshal(jsonBytes, &service)
	if err != nil {
		klog.Errorf("Error while parsing JSON service from file: %s", err.Error())
		return nil, err
	}
	return &service, nil
}
