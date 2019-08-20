package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithDockerImageFlag", func() {
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
		transformedManifest, executeErr = TransformManifestWithDockerImageFlag(originalManifest, overrides)
	})

	When("docker image flag is set", func() {
		When("there is a single app in the manifest with a docker image specified", func() {
			BeforeEach(func() {
				overrides.DockerImage = "some-docker-image"

				originalManifest.Applications = []manifestparser.Application{
					{
						ApplicationModel: manifestparser.ApplicationModel{
							Docker: &manifestparser.Docker{Image: "old-docker-image"},
						},
					},
				}
			})

			It("will override the docker image in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications[0].Docker.Image).To(Equal("some-docker-image"))
			})
		})

		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				overrides.DockerImage = "some-docker-image"

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
