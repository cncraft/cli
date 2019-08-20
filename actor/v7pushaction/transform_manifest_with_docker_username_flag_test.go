package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithDockerUsernameFlag", func() {
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
		transformedManifest, executeErr = TransformManifestWithDockerUsernameFlag(originalManifest, overrides)
	})

	When("docker username flag is set", func() {
		When("there is a single app in the manifest with a docker username specified", func() {
			BeforeEach(func() {
				overrides.DockerUsername = "some-docker-username"

				originalManifest.Applications = []manifestparser.Application{
					{
						ApplicationModel: manifestparser.ApplicationModel{
							Docker: &manifestparser.Docker{Username: "old-docker-username"},
						},
					},
				}
			})

			It("will override the docker username in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications[0].Docker.Username).To(Equal("some-docker-username"))
			})
		})

		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				overrides.DockerUsername = "some-docker-username"

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
