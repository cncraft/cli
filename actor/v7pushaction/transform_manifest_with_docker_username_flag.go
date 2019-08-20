package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func TransformManifestWithDockerUsernameFlag(manifest manifestparser.ParsedManifest, overrides FlagOverrides) (manifestparser.ParsedManifest, error) {
	if overrides.DockerUsername != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}
		manifest.Applications[0].Docker.Username = overrides.DockerUsername
	}

	return manifest, nil
}
