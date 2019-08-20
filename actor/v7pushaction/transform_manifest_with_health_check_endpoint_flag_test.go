package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithHealthCheckEndpointFlag", func() {
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
		transformedManifest, executeErr = TransformManifestWithHealthCheckEndpointFlag(originalManifest, overrides)
	})

	When("manifest web process does not specify health check type", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "web"},
						},
					},
				},
			}
		})

		When("health check type is not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest).To(Equal(originalManifest))
			})
		})

		When("health check type set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.HealthCheckEndpoint = "/health"
			})

			It("changes the health check type of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				process := transformedManifest.Applications[0].Processes[0]
				Expect(process.HealthCheckEndpoint).To(Equal("/health"))
			})
		})
	})

	When("health check type flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckEndpoint = "/health"

			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "worker", HealthCheckEndpoint: "/health2"},
						},
					},
				},
			}
		})

		It("changes the health check type in the app level only", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			Expect(workerProcess.HealthCheckEndpoint).To(Equal("/health2"))
			Expect(transformedManifest.Applications[0].HealthCheckEndpoint).To(Equal("/health"))
		})
	})

	When("health check type flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckEndpoint = "/health"

			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "worker", HealthCheckEndpoint: "/health2"},
							{Type: "web", HealthCheckEndpoint: "/health3"},
						},
						HealthCheckEndpoint: "/health2",
					},
				},
			}
		})

		It("changes the health check type of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			webProcess := transformedManifest.Applications[0].Processes[1]
			Expect(workerProcess.HealthCheckEndpoint).To(Equal("/health2"))
			Expect(webProcess.HealthCheckEndpoint).To(Equal("/health"))
			Expect(transformedManifest.Applications[0].HealthCheckEndpoint).To(Equal("/health2"))
		})
	})

	When("health check type flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.HealthCheckEndpoint = "/health"

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
