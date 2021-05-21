package config

import (
	"fmt"
	"testing"
)

const (
	testConfigFile = "test_config.json"
)

func TestReadConfig(t *testing.T) {
	c, err := ReadConfig(testConfigFile)
	if err != nil {
		t.Errorf("Error while trying to read %s: %s", testConfigFile, err.Error())
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
}
