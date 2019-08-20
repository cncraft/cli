package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithHealthCheckTypeFlag", func() {
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
		transformedManifest, executeErr = TransformManifestWithHealthCheckTypeFlag(originalManifest, overrides)
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
				overrides.HealthCheckType = constant.HTTP
			})

			It("changes the health check type of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				process := transformedManifest.Applications[0].Processes[0]
				Expect(process.HealthCheckType).To(Equal(constant.HTTP))
			})
		})
	})

	When("health check type flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckType = constant.HTTP

			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "worker", HealthCheckType: constant.Port},
						},
					},
				},
			}
		})

		It("changes the health check type in the app level only", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			Expect(workerProcess.HealthCheckType).To(Equal(constant.Port))
			Expect(transformedManifest.Applications[0].HealthCheckType).To(Equal(constant.HTTP))
		})
	})

	When("health check type flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckType = constant.HTTP

			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "worker", HealthCheckType: constant.Port},
							{Type: "web", HealthCheckType: constant.Process},
						},
						HealthCheckType: constant.Port,
					},
				},
			}
		})

		It("changes the health check type of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			webProcess := transformedManifest.Applications[0].Processes[1]
			Expect(workerProcess.HealthCheckType).To(Equal(constant.Port))
			Expect(webProcess.HealthCheckType).To(Equal(constant.HTTP))
			Expect(transformedManifest.Applications[0].HealthCheckType).To(Equal(constant.Port))
		})
	})

	When("health check type flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.HealthCheckType = constant.HTTP

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
