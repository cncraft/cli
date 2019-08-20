package manifestparser

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

// ApplicationModel can be accessed through the top level Application struct To
// add a field for the CLI to extract from the manifest, just add it to this
// struct.
type ApplicationModel struct {
	StartCommand        string                   `yaml:"command,omitempty"`
	Name                string                   `yaml:"name"`
	Buildpacks          []string                 `yaml:"buildpacks,omitempty"`
	DiskQuota           string                   `yaml:"disk_quota,omitempty"`
	Docker              *Docker                  `yaml:"docker,omitempty"`
	HealthCheckType     constant.HealthCheckType `yaml:"health-check-type,omitempty"`
	HealthCheckEndpoint string                   `yaml:"health-check-http-endpoint,omitempty"`
	HealthCheckTimeout  int64                    `yaml:"health-check-invocation-timeout,omitempty"`
	Instances           int                      `yaml:"instances,omitempty"`
	Path                string                   `yaml:"path,omitempty"`
	Processes           []ProcessModel           `yaml:"processes,omitempty"`
	Memory              string                   `yaml:"memory,omitempty"`
	NoRoute             bool                     `yaml:"no-route,omitempty"`
	RandomRoute         bool                     `yaml:"random-route,omitempty"`
	Stack               string                   `yaml:"stack,omitempty"`
}

type Docker struct {
	Image    string `yaml:"image,omitempty"`
	Username string `yaml:"username,omitempty"`
}

type ProcessModel struct {
	StartCommand        string                   `yaml:"command,omitempty"`
	DiskQuota           string                   `yaml:"disk_quota,omitempty"`
	HealthCheckEndpoint string                   `yaml:"health-check-http-endpoint,omitempty"`
	HealthCheckType     constant.HealthCheckType `yaml:"health-check-type,omitempty"`
	HealthCheckTimeout  int64                    `yaml:"health-check-invocation-timeout,omitempty"`
	Instances           int                      `yaml:"instances,omitempty"`
	Memory              string                   `yaml:"memory,omitempty"`
	Type                string                   `yaml:"type"`
}

type Application struct {
	ApplicationModel
	//FullUnmarshalledApplication map[string]interface{}
}

func (application Application) MarshalYAML() (interface{}, error) {
	return application.ApplicationModel, nil
}

func (application *Application) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	//err := unmarshal(&application.FullUnmarshalledApplication)
	//if err != nil {
	//	return err
	//}
	return unmarshal(&application.ApplicationModel)
}
