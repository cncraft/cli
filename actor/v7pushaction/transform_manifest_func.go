package v7pushaction

import "code.cloudfoundry.org/cli/util/manifestparser"

type TransformManifestFunc func(manifest manifestparser.ParsedManifest, overrides FlagOverrides) (manifestparser.ParsedManifest, error)
