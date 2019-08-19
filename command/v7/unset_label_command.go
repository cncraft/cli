package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . UnsetLabelActor

type UnsetLabelActor interface {
	UpdateApplicationLabelsByApplicationName(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateBuildpackLabelsByBuildpackNameAndStack(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateOrganizationLabelsByOrganizationName(string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateSpaceLabelsBySpaceName(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateStackLabelsByStackName(string, map[string]types.NullString) (v7action.Warnings, error)
}

type UnsetLabelCommand struct {
	RequiredArgs   flag.UnsetLabelArgs `positional-args:"yes"`
	BuildpackStack string              `long:"stack" short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`
	usage          interface{}         `usage:"CF_NAME unset-label RESOURCE RESOURCE_NAME KEY\n\nEXAMPLES:\n   cf unset-label app dora ci_signature_sha2\n\nRESOURCES:\n   app\n   buildpack\n   org\n   space\n   stack\n\nSEE ALSO:\n   set-label, labels"`
	UI             command.UI
	Config         command.Config
	SharedActor    command.SharedActor
	Actor          UnsetLabelActor
}

func (cmd *UnsetLabelCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)
	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())
	return nil
}

func (cmd UnsetLabelCommand) Execute(args []string) error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	err = cmd.validateFlags()
	if err != nil {
		return err
	}

	labels := make(map[string]types.NullString)
	for _, value := range cmd.RequiredArgs.LabelKeys {
		labels[value] = types.NewNullString()
	}

	resourceTypeString := strings.ToLower(cmd.RequiredArgs.ResourceType)
	switch ResourceType(resourceTypeString) {
	case App:
		err = cmd.executeApp(user.Name, labels)
	case Buildpack:
		err = cmd.executeBuildpack(user.Name, labels)
	case Org:
		err = cmd.executeOrg(user.Name, labels)
	case Space:
		err = cmd.executeSpace(user.Name, labels)
	case Stack:
		err = cmd.executeStack(user.Name, labels)
	default:
		err = errors.New(cmd.UI.TranslateText("Unsupported resource type of '{{.ResourceType}}'", map[string]interface{}{"ResourceType": cmd.RequiredArgs.ResourceType}))
	}

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd UnsetLabelCommand) executeApp(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Removing label(s) for app {{.ResourceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"SpaceName":    cmd.Config.TargetedSpace().Name,
		"User":         username,
	})

	warnings, err := cmd.Actor.UpdateApplicationLabelsByApplicationName(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedSpace().GUID, labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeBuildpack(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	var template string
	if cmd.BuildpackStack == "" {
		template = "Removing label(s) for buildpack {{.ResourceName}} as {{.User}}..."
	} else {
		template = "Removing label(s) for buildpack {{.ResourceName}} with stack {{.StackName}} as {{.User}}..."
	}
	cmd.UI.DisplayTextWithFlavor(template, map[string]interface{}{
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"StackName":    cmd.BuildpackStack,
		"User":         username,
	})

	warnings, err := cmd.Actor.UpdateBuildpackLabelsByBuildpackNameAndStack(cmd.RequiredArgs.ResourceName, cmd.BuildpackStack, labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeOrg(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Removing label(s) for org {{.ResourceName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"User":         username,
	})

	warnings, err := cmd.Actor.UpdateOrganizationLabelsByOrganizationName(cmd.RequiredArgs.ResourceName, labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeSpace(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Removing label(s) for space {{.ResourceName}} in org {{.OrgName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"User":         username,
	})

	warnings, err := cmd.Actor.UpdateSpaceLabelsBySpaceName(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedOrganization().GUID, labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeStack(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Removing label(s) for stack {{.ResourceName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"User":         username,
	})

	warnings, err := cmd.Actor.UpdateStackLabelsByStackName(cmd.RequiredArgs.ResourceName, labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) validateFlags() error {
	resourceTypeString := strings.ToLower(cmd.RequiredArgs.ResourceType)
	if cmd.BuildpackStack != "" && ResourceType(resourceTypeString) != Buildpack {
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				cmd.RequiredArgs.ResourceType, "--stack, -s",
			},
		}
	}
	return nil
}
