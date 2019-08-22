package v7

import (
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/clock"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	v6shared "code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/util/manifestparser"
	"code.cloudfoundry.org/cli/util/progressbar"

	"github.com/cloudfoundry/bosh-cli/director/template"
	log "github.com/sirupsen/logrus"
)

//go:generate counterfeiter . ProgressBar

type ProgressBar interface {
	v7pushaction.ProgressBar
	Complete()
	Ready()
}

//go:generate counterfeiter . PushActor

type PushActor interface {
	TransformManifest(baseManifest manifestparser.ParsedManifest, flagOverrides v7pushaction.FlagOverrides) (manifestparser.ParsedManifest, error)
	CreatePushPlans(appNameArg string, spaceGUID string, orgGUID string, parser v7pushaction.ManifestParser, overrides v7pushaction.FlagOverrides) ([]v7pushaction.PushPlan, error)
	// Prepare the space by creating needed apps/applying the manifest
	PrepareSpace(pushPlans []v7pushaction.PushPlan, parser v7pushaction.ManifestParser) ([]string, <-chan *v7pushaction.PushEvent)
	// UpdateApplicationSettings figures out the state of the world.
	UpdateApplicationSettings(pushPlans []v7pushaction.PushPlan) ([]v7pushaction.PushPlan, v7pushaction.Warnings, error)
	// Actualize applies any necessary changes.
	Actualize(plan v7pushaction.PushPlan, progressBar v7pushaction.ProgressBar) <-chan *v7pushaction.PushEvent
}

//go:generate counterfeiter . V7ActorForPush

type V7ActorForPush interface {
	AppActor
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v7action.NOAAClient) (<-chan *v7action.LogMessage, <-chan error, v7action.Warnings, error)
	RestartApplication(appGUID string, noWait bool) (v7action.Warnings, error)
}

//go:generate counterfeiter . ManifestParser

type ManifestParser interface {
	v7pushaction.ManifestParser
	GetParsedManifest() manifestparser.ParsedManifest
	ContainsMultipleApps() bool
	InterpolateAndParse(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV, appName string) error
	ContainsPrivateDockerImages() bool
}

//go:generate counterfeiter . ManifestLocator

type ManifestLocator interface {
	Path(filepathOrDirectory string) (string, bool, error)
}

type PushCommand struct {
	OptionalArgs            flag.OptionalAppName                `positional-args:"yes"`
	HealthCheckTimeout      flag.PositiveInteger                `long:"app-start-timeout" short:"t" description:"Time (in seconds) allowed to elapse between starting up an app and the first healthy response from the app"`
	Buildpacks              []string                            `long:"buildpack" short:"b" description:"Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'"`
	Disk                    string                              `long:"disk" short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	DockerImage             flag.DockerImage                    `long:"docker-image" short:"o" description:"Docker image to use (e.g. user/docker-image-name)"`
	DockerUsername          string                              `long:"docker-username" description:"Repository username; used with password from environment variable CF_DOCKER_PASSWORD"`
	DropletPath             flag.PathWithExistenceCheck         `long:"droplet" description:"Path to a tgz file with a pre-staged app"`
	HealthCheckHTTPEndpoint string                              `long:"endpoint"  description:"Valid path on the app for an HTTP health check. Only used when specifying --health-check-type=http"`
	HealthCheckType         flag.HealthCheckType                `long:"health-check-type" short:"u" description:"Application health check type. Defaults to 'port'. 'http' requires a valid endpoint, for example, '/health'."`
	Instances               flag.Instances                      `long:"instances" short:"i" description:"Number of instances"`
	PathToManifest          flag.ManifestPathWithExistenceCheck `long:"manifest" short:"f" description:"Path to manifest"`
	Memory                  string                              `long:"memory" short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	NoManifest              bool                                `long:"no-manifest" description:""`
	NoRoute                 bool                                `long:"no-route" description:"Do not map a route to this app"`
	NoStart                 bool                                `long:"no-start" description:"Do not stage and start the app after pushing"`
	NoWait                  bool                                `long:"no-wait" description:"Do not wait for the long-running operation to complete; push exits when one instance of the web process is healthy"`
	AppPath                 flag.PathWithExistenceCheck         `long:"path" short:"p" description:"Path to app directory or to a zip file of the contents of the app directory"`
	RandomRoute             bool                                `long:"random-route" description:"Create a random route for this app (except when no-route is specified in the manifest)"`
	Stack                   string                              `long:"stack" short:"s" description:"Stack to use (a stack is a pre-built file system, including an operating system, that can run apps)"`
	StartCommand            flag.Command                        `long:"start-command" short:"c" description:"Startup command, set to null to reset to default start command"`
	Strategy                flag.DeploymentStrategy             `long:"strategy" description:"Deployment strategy, either rolling or null."`
	Vars                    []template.VarKV                    `long:"var" description:"Variable key value pair for variable substitution, (e.g., name=app1); can specify multiple times"`
	PathsToVarsFiles        []flag.PathWithExistenceCheck       `long:"vars-file" description:"Path to a variable substitution file for manifest; can specify multiple times"`
	dockerPassword          interface{}                         `environmentName:"CF_DOCKER_PASSWORD" environmentDescription:"Password used for private docker repository"`
	usage                   interface{}                         `usage:"CF_NAME push APP_NAME [-b BUILDPACK_NAME] [-c COMMAND]\n   [-f MANIFEST_PATH | --no-manifest] [--no-start] [--no-wait] [-i NUM_INSTANCES]\n   [-k DISK] [-m MEMORY] [-p PATH] [-s STACK] [-t HEALTH_TIMEOUT]\n   [-u (process | port | http)]   [--no-route | --random-route]\n   [--var KEY=VALUE] [--vars-file VARS_FILE_PATH]...\n \n  CF_NAME push APP_NAME --docker-image [REGISTRY_HOST:PORT/]IMAGE[:TAG] [--docker-username USERNAME]\n   [-c COMMAND] [-f MANIFEST_PATH | --no-manifest] [--no-start] [--no-wait]\n   [-i NUM_INSTANCES] [-k DISK] [-m MEMORY] [-p PATH] [-s STACK] [-t HEALTH_TIMEOUT] [-u (process | port | http)]\n   [--no-route | --random-route ] [--var KEY=VALUE] [--vars-file VARS_FILE_PATH]..."`
	envCFStagingTimeout     interface{}                         `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout     interface{}                         `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	Config          command.Config
	UI              command.UI
	NOAAClient      v3action.NOAAClient
	Actor           PushActor
	VersionActor    V7ActorForPush
	SharedActor     command.SharedActor
	ProgressBar     ProgressBar
	PWD             string
	ManifestLocator ManifestLocator
	ManifestParser  ManifestParser
}

func (cmd *PushCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.ProgressBar = progressbar.NewProgressBar()

	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}

	v7actor := v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	cmd.VersionActor = v7actor
	cmd.Actor = v7pushaction.NewActor(v7actor, sharedActor)

	cmd.NOAAClient = v6shared.NewNOAAClient(ccClient.Info.Logging(), config, uaaClient, ui)

	currentDir, err := os.Getwd()
	cmd.PWD = currentDir

	cmd.ManifestLocator = manifestparser.NewLocator()
	cmd.ManifestParser = manifestparser.NewParser()

	return err
}

func (cmd PushCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	// GET BASE MANIFEST (either read from disk, or start with new empty manifest)
	baseManifest, err := cmd.GetBaseManifest()
	if err != nil {
		return err
	}

	flagOverrides, err := cmd.GetFlagOverrides()
	if err != nil {
		return err
	}

	err = cmd.ValidateFlags()
	if err != nil {
		return err
	}

	// TRANSFORM MANIFEST
	//   start with the base manifest
	//   for each flag, create a function that transforms the manifest according to the flag
	//   end with the transformed manifest
	transformedManifest, err := cmd.Actor.TransformManifest(baseManifest, flagOverrides)
	if err != nil {
		return err
	}

	// APPLY MANIFEST (hit /apply_manifest endpoint with transformed manifest)
	err = cmd.Actor.ApplySpaceManifest(transformedManifest.RawManifest())
	if err != nil {
		return err
	}

	// DONE FOR NOW

	//flagOverrides.DockerPassword, err = cmd.GetDockerPassword(flagOverrides.DockerUsername, cmd.ManifestParser.ContainsPrivateDockerImages())
	//if err != nil {
	//	return err
	//}
	//
	//pushPlans, err := cmd.Actor.CreatePushPlans(
	//	cmd.OptionalArgs.AppName,
	//	cmd.Config.TargetedSpace().GUID,
	//	cmd.Config.TargetedOrganization().GUID,
	//	cmd.ManifestParser,
	//	flagOverrides,
	//)
	//if err != nil {
	//	return err
	//}
	//
	//appNames, eventStream := cmd.Actor.PrepareSpace(pushPlans, cmd.ManifestParser)
	//err = cmd.eventStreamHandler(eventStream)
	//
	//if err != nil {
	//	return err
	//}
	//
	//if len(appNames) == 0 {
	//	return translatableerror.AppNameOrManifestRequiredError{}
	//}
	//
	//user, err := cmd.Config.CurrentUser()
	//if err != nil {
	//	return err
	//}
	//
	//cmd.announcePushing(appNames, user)
	//
	//cmd.UI.DisplayText("Getting app info...")
	//log.Info("generating the app plan")
	//
	//pushPlans, warnings, err := cmd.Actor.UpdateApplicationSettings(pushPlans)
	//cmd.UI.DisplayWarnings(warnings)
	//if err != nil {
	//	return err
	//}
	//log.WithField("number of plans", len(pushPlans)).Debug("completed generating plan")
	//
	//for _, plan := range pushPlans {
	//	log.WithField("app_name", plan.Application.Name).Info("actualizing")
	//	eventStream := cmd.Actor.Actualize(plan, cmd.ProgressBar)
	//	err := cmd.eventStreamHandler(eventStream)
	//
	//	if cmd.shouldDisplaySummary(err) {
	//		summaryErr := cmd.displayAppSummary(plan)
	//		if summaryErr != nil {
	//			return summaryErr
	//		}
	//	}
	//	if err != nil {
	//		return cmd.mapErr(plan.Application.Name, err)
	//	}
	//}

	return nil
}

func (cmd PushCommand) GetBaseManifest() (manifestparser.ParsedManifest, error) {
	// if no-manifest flag is passed
	//   return an empty manifest
	if cmd.NoManifest {
		return cmd.ManifestParser.GetParsedManifest(), nil
	}

	// read manifest file from disk & parse it (if it exists)
	err := cmd.ReadManifest()
	if err != nil {
		return manifestparser.ParsedManifest{}, err
	}

	// return parsed manifest.yml
	return cmd.ManifestParser.GetParsedManifest(), nil
}

func (cmd PushCommand) shouldDisplaySummary(err error) bool {
	if err == nil {
		return true
	}
	_, ok := err.(actionerror.AllInstancesCrashedError)
	return ok
}

func (cmd PushCommand) mapErr(appName string, err error) error {
	switch err.(type) {
	case actionerror.AllInstancesCrashedError:
		return translatableerror.ApplicationUnableToStartError{
			AppName:    appName,
			BinaryName: cmd.Config.BinaryName(),
		}
	case actionerror.StartupTimeoutError:
		return translatableerror.StartupTimeoutError{
			AppName:    appName,
			BinaryName: cmd.Config.BinaryName(),
		}
	}
	return err
}

func (cmd PushCommand) announcePushing(appNames []string, user configv3.User) {
	tokens := map[string]interface{}{
		"AppName":   strings.Join(appNames, ", "),
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	}
	singular := "Pushing app {{.AppName}} to org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}..."
	plural := "Pushing apps {{.AppName}} to org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}..."

	if len(appNames) == 1 {
		cmd.UI.DisplayTextWithFlavor(singular, tokens)
	} else {
		cmd.UI.DisplayTextWithFlavor(plural, tokens)
	}
}

func (cmd PushCommand) displayAppSummary(plan v7pushaction.PushPlan) error {
	log.Info("getting application summary info")
	summary, warnings, err := cmd.VersionActor.GetDetailedAppSummary(
		plan.Application.Name,
		cmd.Config.TargetedSpace().GUID,
		true,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayNewline()
	appSummaryDisplayer := shared.NewAppSummaryDisplayer(cmd.UI)
	appSummaryDisplayer.AppDisplay(summary, true)
	return nil
}

func (cmd PushCommand) eventStreamHandler(eventStream <-chan *v7pushaction.PushEvent) error {
	for event := range eventStream {
		cmd.UI.DisplayWarnings(event.Warnings)
		if event.Err != nil {
			return event.Err
		}
		_, err := cmd.processEvent(event.Event, event.Plan.Application.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cmd PushCommand) processEvent(event v7pushaction.Event, appName string) (bool, error) {
	switch event {
	case v7pushaction.SkippingApplicationCreation:
		cmd.UI.DisplayTextWithFlavor("Updating app {{.AppName}}...", map[string]interface{}{
			"AppName": appName,
		})
	case v7pushaction.CreatingApplication:
		cmd.UI.DisplayTextWithFlavor("Creating app {{.AppName}}...", map[string]interface{}{
			"AppName": appName,
		})
	case v7pushaction.CreatingAndMappingRoutes:
		cmd.UI.DisplayText("Mapping routes...")
	case v7pushaction.CreatingArchive:
		cmd.UI.DisplayText("Packaging files to upload...")
	case v7pushaction.UploadingApplicationWithArchive:
		cmd.UI.DisplayText("Uploading files...")
		log.Debug("starting progress bar")
		cmd.ProgressBar.Ready()
	case v7pushaction.UploadingApplication:
		cmd.UI.DisplayText("All files found in remote cache; nothing to upload.")
		cmd.UI.DisplayText("Waiting for API to complete processing files...")
	case v7pushaction.RetryUpload:
		cmd.UI.DisplayText("Retrying upload due to an error...")
	case v7pushaction.UploadWithArchiveComplete:
		cmd.ProgressBar.Complete()
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Waiting for API to complete processing files...")
	case v7pushaction.UploadingDroplet:
		cmd.UI.DisplayText("Uploading droplet bits...")
		cmd.ProgressBar.Ready()
	case v7pushaction.UploadDropletComplete:
		cmd.ProgressBar.Complete()
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Waiting for API to complete processing files...")
	case v7pushaction.StoppingApplication:
		cmd.UI.DisplayText("Stopping Application...")
	case v7pushaction.StoppingApplicationComplete:
		cmd.UI.DisplayText("Application Stopped")
	case v7pushaction.ApplyManifest:
		cmd.UI.DisplayText("Applying manifest...")
	case v7pushaction.ApplyManifestComplete:
		cmd.UI.DisplayText("Manifest applied")
	case v7pushaction.StartingStaging:
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Staging app and tracing logs...")
		logStream, errStream, warnings, err := cmd.VersionActor.GetStreamingLogsForApplicationByNameAndSpace(appName, cmd.Config.TargetedSpace().GUID, cmd.NOAAClient)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return false, err
		}
		go cmd.getLogs(logStream, errStream)
	case v7pushaction.StagingComplete:
		cmd.NOAAClient.Close()
	case v7pushaction.RestartingApplication:
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayTextWithFlavor(
			"Waiting for app {{.AppName}} to start...",
			map[string]interface{}{
				"AppName": appName,
			},
		)
	case v7pushaction.StartingDeployment:
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayTextWithFlavor(
			"Starting deployment for app {{.AppName}}...",
			map[string]interface{}{
				"AppName": appName,
			},
		)
	case v7pushaction.WaitingForDeployment:
		cmd.UI.DisplayText("Waiting for app to deploy...")
	case v7pushaction.Complete:
		return true, nil
	default:
		log.WithField("event", event).Debug("ignoring event")
	}
	return false, nil
}

func (cmd PushCommand) getLogs(logStream <-chan *v7action.LogMessage, errStream <-chan error) {
	for {
		select {
		case logMessage, open := <-logStream:
			if !open {
				return
			}
			if logMessage.Staging() {
				cmd.UI.DisplayLogMessage(logMessage, false)
			}
		case err, open := <-errStream:
			if !open {
				return
			}
			_, ok := err.(actionerror.NOAATimeoutError)
			if ok {
				cmd.UI.DisplayWarning("timeout connecting to log server, no log will be shown")
			}
			cmd.UI.DisplayWarning(err.Error())
		}
	}
}

func (cmd PushCommand) ReadManifest() error {
	log.Info("reading manifest if exists")
	pathsToVarsFiles := []string{}
	for _, varfilepath := range cmd.PathsToVarsFiles {
		pathsToVarsFiles = append(pathsToVarsFiles, string(varfilepath))
	}

	readPath := cmd.PWD
	if len(cmd.PathToManifest) != 0 {
		log.WithField("manifestPath", cmd.PathToManifest).Debug("reading '-f' provided manifest")
		readPath = string(cmd.PathToManifest)
	}

	pathToManifest, exists, err := cmd.ManifestLocator.Path(readPath)
	if err != nil {
		return err
	}

	if exists {
		log.WithField("manifestPath", pathToManifest).Debug("path to manifest")
		err = cmd.ManifestParser.InterpolateAndParse(pathToManifest, pathsToVarsFiles, cmd.Vars, cmd.OptionalArgs.AppName)
		if err != nil {
			log.Errorln("reading manifest:", err)
			return err
		}

		cmd.UI.DisplayText("Using manifest file {{.Path}}", map[string]interface{}{"Path": pathToManifest})
	}

	return nil
}

func (cmd PushCommand) GetFlagOverrides() (v7pushaction.FlagOverrides, error) {
	return v7pushaction.FlagOverrides{
		AppName:             cmd.OptionalArgs.AppName,
		Buildpacks:          cmd.Buildpacks,
		Stack:               cmd.Stack,
		Disk:                cmd.Disk,
		DropletPath:         string(cmd.DropletPath),
		DockerImage:         cmd.DockerImage.Path,
		DockerUsername:      cmd.DockerUsername,
		HealthCheckEndpoint: cmd.HealthCheckHTTPEndpoint,
		HealthCheckType:     cmd.HealthCheckType.Type,
		HealthCheckTimeout:  cmd.HealthCheckTimeout.Value,
		Instances:           cmd.Instances.NullInt,
		Memory:              cmd.Memory,
		NoStart:             cmd.NoStart,
		NoWait:              cmd.NoWait,
		ProvidedAppPath:     string(cmd.AppPath),
		NoRoute:             cmd.NoRoute,
		RandomRoute:         cmd.RandomRoute,
		StartCommand:        cmd.StartCommand.FilteredString,
		Strategy:            cmd.Strategy.Name,
	}, nil
}

func (cmd PushCommand) ValidateAllowedFlagsForMultipleApps(containsMultipleApps bool) error {
	if cmd.OptionalArgs.AppName != "" {
		return nil
	}

	allowedFlagsMultipleApps := !(
		cmd.DropletPath != "" ||
		cmd.AppPath != "" ||
		cmd.Strategy.Name != "")

	if containsMultipleApps && !allowedFlagsMultipleApps {
		return translatableerror.CommandLineArgsWithMultipleAppsError{}
	}

	return nil
}

func (cmd PushCommand) ValidateFlags() error {
	switch {
	case cmd.DockerUsername != "" && cmd.DockerImage.Path == "":
		return translatableerror.RequiredFlagsError{
			Arg1: "--docker-image, -o",
			Arg2: "--docker-username",
		}

	case cmd.DockerImage.Path != "" && cmd.Buildpacks != nil:
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--buildpack, -b",
				"--docker-image, -o",
			},
		}

	case cmd.DockerImage.Path != "" && cmd.AppPath != "":
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--docker-image, -o",
				"--path, -p",
			},
		}

	case cmd.DockerImage.Path != "" && cmd.Stack != "":
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--stack, -s",
				"--docker-image, -o",
			},
		}

	case cmd.NoManifest && cmd.PathToManifest != "":
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--no-manifest",
				"--manifest, -f",
			},
		}

	case cmd.NoManifest && len(cmd.PathsToVarsFiles) > 0:
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--no-manifest",
				"--vars-file",
			},
		}

	case cmd.NoManifest && len(cmd.Vars) > 0:
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--no-manifest",
				"--vars",
			},
		}

	case cmd.HealthCheckType.Type == constant.HTTP && cmd.HealthCheckHTTPEndpoint == "":
		return translatableerror.RequiredFlagsError{
			Arg1: "--endpoint",
			Arg2: "--health-check-type=http, -u=http",
		}

	case 0 < len(cmd.HealthCheckHTTPEndpoint) && cmd.HealthCheckType.Type != constant.HTTP:
		return translatableerror.RequiredFlagsError{
			Arg1: "--health-check-type=http, -u=http",
			Arg2: "--endpoint",
		}

	case cmd.DropletPath != "" && (cmd.DockerImage.Path != "" || cmd.DockerUsername != "" || cmd.AppPath != ""):
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--droplet",
				"--docker-image, -o",
				"--docker-username",
				"-p",
			},
		}

	case cmd.NoStart && cmd.Strategy == flag.DeploymentStrategy{Name: constant.DeploymentStrategyRolling}:
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--no-start",
				"--strategy=rolling",
			},
		}

	case cmd.NoStart && cmd.NoWait:
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--no-start",
				"--no-wait",
			},
		}

	case cmd.NoRoute && cmd.RandomRoute:
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				"--no-route",
				"--random-route",
			},
		}

	}

	return nil
}
