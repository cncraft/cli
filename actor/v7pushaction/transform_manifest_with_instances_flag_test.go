package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithInstancesFlag", func() {
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
		transformedManifest, executeErr = TransformManifestWithInstancesFlag(originalManifest, overrides)
	})

	When("manifest web process does not specify instances", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "web", StartCommand: "./start.sh"},
						},
					},
				},
			}
		})

		When("instances are not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest).To(Equal(originalManifest))
			})
		})

		When("instances are set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.Instances = types.NullInt{IsSet: true, Value: 4}
			})

			It("changes the instances of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				process := transformedManifest.Applications[0].Processes[0]
				Expect(process.Instances).To(Equal(4))
			})
		})
	})

	When("instances flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.Instances = types.NullInt{IsSet: true, Value: 4}

			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "worker", StartCommand: "./work.sh"},
						},
					},
				},
			}
		})

		It("changes the instances of the app in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			Expect(workerProcess.Instances).To(Equal(0))
			Expect(transformedManifest.Applications[0].Instances).To(Equal(4))
		})
	})

	When("instances flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.Instances = types.NullInt{IsSet: true, Value: 4}

			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "worker", StartCommand: "./work.sh"},
							{Type: "web", StartCommand: "./start.sh"},
						},
						Instances: 5,
					},
				},
			}
		})

		It("changes the instances of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			webProcess := transformedManifest.Applications[0].Processes[1]
			Expect(workerProcess.Instances).To(Equal(0))
			Expect(webProcess.Instances).To(Equal(4))
			Expect(transformedManifest.Applications[0].Instances).To(Equal(5))

		})
	})

	When("instances flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.Instances = types.NullInt{IsSet: true, Value: 4}

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
