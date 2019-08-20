package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func TransformManifestWithDockerImageFlag(manifest manifestparser.ParsedManifest, overrides FlagOverrides) (manifestparser.ParsedManifest, error) {
	if overrides.DockerImage != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}
		manifest.Applications[0].Docker.Image = overrides.DockerImage
	}

	return manifest, nil
}
