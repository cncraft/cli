package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func TransformManifestWithHealthCheckTypeFlag(manifest manifestparser.ParsedManifest, overrides FlagOverrides) (manifestparser.ParsedManifest, error) {
	if overrides.HealthCheckType != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.HealthCheckType = overrides.HealthCheckType
		} else {
			app := manifest.GetFirstApp()
			app.HealthCheckType = overrides.HealthCheckType
		}
	}

	return manifest, nil
}
