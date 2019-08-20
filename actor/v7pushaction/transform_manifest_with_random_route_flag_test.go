package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithRandomRouteFlag", func() {
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
		transformedManifest, executeErr = TransformManifestWithRandomRouteFlag(originalManifest, overrides)
	})

	When("manifest app does not specify random-route", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						NoRoute: true,
					},
				},
			}
		})

		When("random-route is not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest).To(Equal(originalManifest))
			})
		})

		When("random-route is set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.RandomRoute = true
			})

			It("changes the random-route field of the only app in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				app := transformedManifest.Applications[0]
				Expect(app.RandomRoute).To(BeTrue())
				Expect(app.NoRoute).To(BeTrue())
			})
		})
	})

	When("random-route flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.RandomRoute = true

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
