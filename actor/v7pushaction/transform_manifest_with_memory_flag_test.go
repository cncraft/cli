package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithMemoryFlag", func() {
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
		transformedManifest, executeErr = TransformManifestWithMemoryFlag(originalManifest, overrides)
	})

	When("manifest web process does not specify memory", func() {
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

		When("memory are not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest).To(Equal(originalManifest))
			})
		})

		When("memory are set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.Memory = "64M"
			})

			It("changes the memory of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				process := transformedManifest.Applications[0].Processes[0]
				Expect(process.Memory).To(Equal("64M"))
			})
		})
	})

	When("memory flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.Memory = "64M"

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

		It("changes the memory of the app in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			Expect(workerProcess.Memory).To(Equal(""))
			Expect(transformedManifest.Applications[0].Memory).To(Equal("64M"))
		})
	})

	When("memory flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.Memory = "64M"

			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "worker", StartCommand: "./work.sh"},
							{Type: "web", StartCommand: "./start.sh"},
						},
						Memory: "8M",
					},
				},
			}
		})

		It("changes the memory of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			webProcess := transformedManifest.Applications[0].Processes[1]
			Expect(workerProcess.Memory).To(Equal(""))
			Expect(webProcess.Memory).To(Equal("64M"))
			Expect(transformedManifest.Applications[0].Memory).To(Equal("8M"))

		})
	})

	When("memory flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.Memory = "64M"

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
