package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithBuildpacksFlag", func() {
	var (
		originalManifest    manifestparser.ParsedManifest
		transformedManifest manifestparser.ParsedManifest
		overrides           FlagOverrides
		executeErr          error
	)

	BeforeEach(func() {
		originalManifest = manifestparser.ParsedManifest{}
		overrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = TransformManifestWithBuildpacksFlag(originalManifest, overrides)
	})

	When("buildpacks flag is set", func() {
		When("there is a single app in the manifest with a buildpacks specified", func() {
			BeforeEach(func() {
				overrides.Buildpacks = []string{"buildpack-1", "buildpack-2"}

				originalManifest.Applications = []manifestparser.Application{
					{
						ApplicationModel: manifestparser.ApplicationModel{
							Buildpacks: []string{"buildpack-3"},
						},
					},
				}
			})

			It("will override the buildpacks in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications[0].Buildpacks).To(ConsistOf("buildpack-1", "buildpack-2"))
			})
		})
		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				overrides.Buildpacks = []string{"buildpack-1", "buildpack-2"}

				originalManifest.Applications = []manifestparser.Application{
					{
						ApplicationModel: manifestparser.ApplicationModel{},
					},
					{
						ApplicationModel: manifestparser.ApplicationModel{},
					},
				}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.CommandLineArgsWithMultipleAppsError{}))
			})
		})
	})
})
