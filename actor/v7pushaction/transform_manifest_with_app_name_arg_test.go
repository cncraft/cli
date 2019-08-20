package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithAppNameArg", func() {
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
		transformedManifest, executeErr = TransformManifestWithAppNameArg(originalManifest, overrides)
	})

	When("manifest web process does not specify instances", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Name: "app-1",
					},
				},
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Name: "app-2",
					},
				},
			}
		})

		When("app name is not given as arg", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest).To(Equal(originalManifest))
			})
		})

		When("a valid app name is set as a manifest override", func() {
			BeforeEach(func() {
				overrides.AppName = "app-2"
			})

			It("removes non-specified apps from manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				numApps := len(transformedManifest.Applications)
				Expect(numApps).To(Equal(1))
				Expect(transformedManifest.Applications[0].Name).To(Equal("app-2"))
			})
		})

		When("there is only one app, with no name, and a name is given as a manifest override", func() {
			BeforeEach(func() {
				originalManifest.Applications = []manifestparser.Application{
					{
						ApplicationModel: manifestparser.ApplicationModel{
						},
					},
				}

				overrides.AppName = "app-2"
			})

			It("gives the app a name in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				numApps := len(transformedManifest.Applications)
				Expect(numApps).To(Equal(1))
				Expect(transformedManifest.Applications[0].Name).To(Equal("app-2"))
			})
		})

		When("there are multiple apps and one does not have a name", func() {
			BeforeEach(func() {
				originalManifest.Applications = []manifestparser.Application{
					{
						ApplicationModel: manifestparser.ApplicationModel{
							Name: "app-1",
						},
					},
					{
						ApplicationModel: manifestparser.ApplicationModel{
							Name: "",
						},
					},
				}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("Found an application with no name specified."))
			})
		})

		When("an invalid app name is set as a manifest override", func() {
			BeforeEach(func() {
				overrides.AppName = "unknown-app"
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(manifestparser.AppNotInManifestError{Name: "unknown-app"}))
			})
		})
	})
})
