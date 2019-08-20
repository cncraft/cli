package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithStackFlag", func() {
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
		transformedManifest, executeErr = TransformManifestWithStackFlag(originalManifest, overrides)
	})

	When("stack flag is set", func() {
		When("there is a single app in the manifest with a stack specified", func() {
			BeforeEach(func() {
				overrides.Stack = "cflinuxfs2"

				originalManifest.Applications = []manifestparser.Application{
					{
						ApplicationModel: manifestparser.ApplicationModel{
							Stack: "cflinuxfs3",
						},
					},
				}
			})

			It("will override the stack in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications[0].Stack).To(Equal("cflinuxfs2"))
			})
		})

		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				overrides.Stack = "cflinuxfs2"

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
