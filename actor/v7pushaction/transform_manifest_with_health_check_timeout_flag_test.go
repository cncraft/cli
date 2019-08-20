package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithHealthCheckTimeoutFlag", func() {
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
		transformedManifest, executeErr = TransformManifestWithHealthCheckTimeoutFlag(originalManifest, overrides)
	})

	When("manifest web process does not specify health check timeout", func() {
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

		When("health check timeout is not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest).To(Equal(originalManifest))
			})
		})

		When("health check timeout set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.HealthCheckTimeout = 50
			})

			It("changes the health check timeout of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				process := transformedManifest.Applications[0].Processes[0]
				Expect(process.HealthCheckTimeout).To(BeEquivalentTo(50))
			})
		})
	})

	When("health check timeout flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckTimeout = 50

			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "worker", HealthCheckTimeout: 10},
						},
					},
				},
			}
		})

		It("changes the health check timeout in the app level only", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			Expect(workerProcess.HealthCheckTimeout).To(BeEquivalentTo(10))
			Expect(transformedManifest.Applications[0].HealthCheckTimeout).To(BeEquivalentTo(50))
		})
	})

	When("health check timeout flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckTimeout = 50

			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "worker", HealthCheckTimeout: 10},
							{Type: "web", HealthCheckTimeout: 20},
						},
						HealthCheckTimeout: 30,
					},
				},
			}
		})

		It("changes the health check timeout of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			webProcess := transformedManifest.Applications[0].Processes[1]
			Expect(workerProcess.HealthCheckTimeout).To(BeEquivalentTo(10))
			Expect(webProcess.HealthCheckTimeout).To(BeEquivalentTo(50))
			Expect(transformedManifest.Applications[0].HealthCheckTimeout).To(BeEquivalentTo(30))
		})
	})

	When("health check timeout flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.HealthCheckTimeout = 50

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
