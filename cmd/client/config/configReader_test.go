package config

import (
	"fmt"
	"testing"
)

const (
	testConfigFile1 = "test_config.json"
	testConfigFile2 = "test_config_service_dir.json"
)

func TestReadConfig(t *testing.T) {
	t.Logf("===== Testing config file %s:", testConfigFile1)
	c, err := ReadConfig(testConfigFile1)
	if err != nil {
		t.Errorf("Error while trying to read %s: %s", testConfigFile1, err.Error())
	}

	if *c.InstanceId != "registry-client_1" {
		t.Errorf("config.InstanceId expected 'registry-client_1' but got %s", *c.InstanceId)
	} else {
		var svcConfDir string = "nil"
		if c.ServiceConfDir != nil {
			svcConfDir = *c.ServiceConfDir
		}
		var r string = "{"
		var serviceName string = "nil"
		var service Service
		for sn, s := range *c.Services {
			serviceName = sn
			service = s
			// we only want the first one, so break immediately
			break
		}

		r = r + fmt.Sprintf("InstanceId: %s, ServiceConfDir: %s, Services:{ServiceName: %s: ", *c.InstanceId, svcConfDir, serviceName)
		r = r + fmt.Sprintf("{Host: %s, Port: %d}}}", *service.Host, *service.Port)
		t.Logf("got result: %s", r)
	}

	t.Logf("===== Testing config file %s:", testConfigFile2)
	c, err = ReadConfig(testConfigFile2)
	if err != nil {
		t.Errorf("Error while trying to read %s: %s", testConfigFile2, err.Error())
	}

	if *c.InstanceId != "registry-client_2" {
		t.Errorf("config.InstanceId expected 'registry-client_2' but got %s", *c.InstanceId)
	} else {
		var svcConfDir string = "nil"
		if c.ServiceConfDir != nil {
			svcConfDir = *c.ServiceConfDir
		}
		var r string = "{"
		noOfServices := len(*c.Services)
		t.Logf("got %d services", noOfServices)

		r = r + fmt.Sprintf("InstanceId: %s, ServiceConfDir: %s, ", *c.InstanceId, svcConfDir)
		r = r + "Services: {"
		var idx int = 0
		for sn, s := range *c.Services {
			r = r + fmt.Sprintf("{ServiceName: %s, Host: %s, Port: %d}", sn, *s.Host, *s.Port)
			idx += 1
			if idx < noOfServices {
				r = r + ","
			} else {
				r = r + "}"
			}
		}
		r = r + "}"
		t.Logf("got result: %s", r)
	}
}
