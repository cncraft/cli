package v7pushaction

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"
)

type PushPlan struct {
	SpaceGUID string
	OrgGUID   string

	Application            v7action.Application
	ApplicationRoutes      []v7action.Route
	ApplicationNeedsUpdate bool

	NoStart           bool
	NoWait            bool
	NoRouteFlag       bool
	RandomRoute       bool
	SkipRouteCreation bool
	Strategy          constant.DeploymentStrategy

	DockerImageCredentials            v7action.DockerImageCredentials
	DockerImageCredentialsNeedsUpdate bool

	ScaleWebProcess            v7action.Process
	ScaleWebProcessNeedsUpdate bool

	UpdateWebProcess            v7action.Process
	UpdateWebProcessNeedsUpdate bool

	Manifest []byte

	Archive      bool
	BitsPath     string
	DropletPath  string
	AllResources []sharedaction.V3Resource

	PackageGUID string
	DropletGUID string
}

type FlagOverrides struct {
	AppName             string
	Buildpacks          []string
	Stack               string
	Disk                types.NullUint64
	DropletPath         string
	DockerImage         string
	DockerPassword      string
	DockerUsername      string
	HealthCheckEndpoint string
	HealthCheckTimeout  int64
	HealthCheckType     constant.HealthCheckType
	Instances           types.NullInt
	Memory              types.NullUint64
	NoStart             bool
	NoWait              bool
	ProvidedAppPath     string
	NoRoute             bool
	RandomRoute         bool
	StartCommand        types.FilteredString
	Strategy            constant.DeploymentStrategy
}

func (state PushPlan) String() string {
	return fmt.Sprintf(
		"Application: %#v - Space GUID: %s, Org GUID: %s, Archive: %t, Bits Path: %s",
		state.Application,
		state.SpaceGUID,
		state.OrgGUID,
		state.Archive,
		state.BitsPath,
	)
}
