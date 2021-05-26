package config

// Helper functions to return a pointer to a variable defined as constant
func StringP(s string) *string          { return &s }
func CheckTypeP(s CheckType) *CheckType { return &s }
func Int32P(i int32) *int32             { return &i }
func Int16P(i int16) *int16             { return &i }

// CheckType - string type defining the check.
type CheckType string

const (
	Tcp  CheckType = "tcp"
	Http CheckType = "http"
	File CheckType = "file"
)

// Check - configuration of a service check
// Type:       the type of the check - see constants above for supported types
// Uri:        the URI used for tcp or http checks, empty string for file
//             for http a full URL is expected such as http(s)://<host>:<port>, e.g. http://localhost:80
//             for tcp the expected format is tcp://<host>:<port>
// FilePath:   the file path for file check, empty string otherwise
// TimeoutMs:  (optional) timeout in milliseconds.
//             connection timeout for tcp check
//             response timeout for http check
//             not applicable for file check
//             a value of 0 means to use the system default which is not recommended - defaults to 100ms
// Rise:       (optional) number of checks to pass before the service is considered healthy - defaults to 1
// Fall:       (optional) number of checks to fail before the service is considered unhealthy - defaults to 1
type Check struct {
	Type      *CheckType `json:"type"`
	Uri       *string    `json:"uri,omitempty"`
	FilePath  *string    `json:"file_path:omitempty"`
	TimeoutMs *int32     `json:"timeout_ms,omitempty"`
	Rise      *int16     `json:"rise,omitempty"`
	Fall      *int16     `json:"fall,omitempty"`
}

// ChecksType - array of Checks
type ChecksType []Check

// Service - configuration of the service to be registered
// Host :           host of the service - this must be either a DNS hostname or IP address
// Port:            port the service is listening on
// Registries:      list of URL of the service registry, e.g. https://service-registry.cluster.local
// CheckIntervalMs: time interval in milliseconds a check is being performed - defaults to 1000ms
//                  IMPORTANT: this value must be longer than the TimeoutMs configured above!
// Checks:          array of  one or many checks that need to be performed. All checks are cummulative,
//                  i.e. all of them have to pass before a service will be registered into the registry as healthy
type Service struct {
	Host            *string     `json:"host"`
	Port            *int32      `json:"port"`
	Registries      []*string   `json:"registries"`
	CheckIntervalMs *int32      `json:"check_interval_ms,omitempty"`
	Checks          *ChecksType `json:"checks"`
}

type ServicesType map[string]Service

// Config - top-level configuration
// InstanceId:     instance ID of that service registry client.
//                 This will be used to identify the client in the service and must be
//                 unique within the entire K8s cluster.
// ServiceConfDir: (optional) path to a folder that contains additional service definitions
// Services:       (optional) map containing service definitions. Does not have to be present!
//                 One of ServiceConfDir with definitions inside or Service with at least one
//                 service definitions must be present.
// IMPORTANT:      The service name will be either the filename if a ServiceConfDir is given, or
//                 the key in the Services map.
type Config struct {
	// Instance ID of that service
	InstanceId     *string       `json:"instance_id"`
	ServiceConfDir *string       `json:"service_conf_dir,omitempty"`
	Services       *ServicesType `json:"services,omitempty"`
}

var DefaultChecks ChecksType = []Check{
	{
		Type:      CheckTypeP("http"),
		Uri:       StringP("/healthz"),
		TimeoutMs: Int32P(500),
		Rise:      Int16P(1),
		Fall:      Int16P(1),
	},
}

var DefaultService Service = Service{
	Host:            StringP("localhost"),
	Port:            Int32P(80),
	Registries:      []*string{StringP("localhost:9080")},
	CheckIntervalMs: Int32P(500),
	Checks:          &DefaultChecks,
}

var EmptyService Service = Service{
	Host:            nil,
	Port:            nil,
	Registries:      nil,
	CheckIntervalMs: nil,
	Checks:          nil,
}

var DefaultConfig Config = Config{
	InstanceId:     StringP(""),
	ServiceConfDir: nil,
	Services:       nil,
}
