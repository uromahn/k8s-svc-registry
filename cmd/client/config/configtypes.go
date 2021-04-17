package config

func StringP(s string) *string          { return &s }
func CheckTypeP(s CheckType) *CheckType { return &s }
func Int32P(i int32) *int32             { return &i }
func Int16P(i int16) *int16             { return &i }

type CheckType string

const (
	Tcp  CheckType = "tcp"
	Http CheckType = "http"
	File CheckType = "file"
)

type Check struct {
	Type      *CheckType `json:"type"`
	Uri       *string    `json:"uri,omitempty"`
	FilePath  *string    `json:"file_path:omitempty"`
	TimeoutMs *int32     `json:"timeout_ms,omitempty"`
	Rise      *int16     `json:"rise,omitempty"`
	Fall      *int16     `json:"fall,omitempty"`
}

type ChecksType []Check

type Service struct {
	Host            *string     `json:"host"`
	Port            *int32      `json:"port"`
	RegistryService *string     `json:"registry_service"`
	CheckIntervalMs *int32      `json:"check_interval_ms,omitempty"`
	Checks          *ChecksType `json:"checks"`
}

type ServicesType map[string]Service

type Config struct {
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
	RegistryService: StringP("localhost:9080"),
	CheckIntervalMs: Int32P(500),
	Checks:          &DefaultChecks,
}

var DefaultConfig Config = Config{
	InstanceId:     StringP(""),
	ServiceConfDir: nil,
	Services:       nil,
}
