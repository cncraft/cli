package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithStartCommandFlag", func() {
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
		transformedManifest, executeErr = TransformManifestWithStartCommandFlag(originalManifest, overrides)
	})

	When("manifest web process does not specify start command", func() {
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

		When("start command is not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest).To(Equal(originalManifest))
			})
		})

		When("start command set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.StartCommand = types.FilteredString{Value: "./start.sh", IsSet: true}
			})

			It("changes the start command of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				process := transformedManifest.Applications[0].Processes[0]
				Expect(process.StartCommand).To(Equal("./start.sh"))
			})
		})
	})

	When("start command flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.StartCommand = types.FilteredString{Value: "./start.sh", IsSet: true}

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

		It("changes the start command in the app level only", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			Expect(workerProcess.StartCommand).To(Equal("./work.sh"))
			Expect(transformedManifest.Applications[0].StartCommand).To(Equal("./start.sh"))
		})
	})

	When("start command flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.StartCommand = types.FilteredString{Value: "./start.sh", IsSet: true}

			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "worker", StartCommand: "./work.sh"},
							{Type: "web", StartCommand: "./web.sh"},
						},
						StartCommand: "./appstart.sh",
					},
				},
			}
		})

		It("changes the start command of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			webProcess := transformedManifest.Applications[0].Processes[1]
			Expect(workerProcess.StartCommand).To(Equal("./work.sh"))
			Expect(webProcess.StartCommand).To(Equal("./start.sh"))
			Expect(transformedManifest.Applications[0].StartCommand).To(Equal("./appstart.sh"))
		})
	})

	When("start command flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.StartCommand = types.FilteredString{Value: "./start.sh", IsSet: true}

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
