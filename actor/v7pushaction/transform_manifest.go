package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func (actor Actor) TransformManifest(
	baseManifest manifestparser.ParsedManifest,
	flagOverrides FlagOverrides,
) (manifestparser.ParsedManifest, error) {
	newManifest := baseManifest

	for _, transformPlan := range actor.TransformManifestSequence {
		var err error
		newManifest, err = transformPlan(newManifest, flagOverrides)
		if err != nil {
			return manifestparser.ParsedManifest{}, err
		}
	}

	return newManifest, nil
}
