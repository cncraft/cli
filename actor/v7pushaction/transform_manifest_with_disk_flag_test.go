package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifestWithDiskFlag", func() {
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
		transformedManifest, executeErr = TransformManifestWithDiskFlag(originalManifest, overrides)
	})

	When("manifest web process does not specify disk", func() {
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

		When("disk is not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest).To(Equal(originalManifest))
			})
		})

		When("disk is set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.Disk = "5MB"
			})

			It("changes the disk of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				process := transformedManifest.Applications[0].Processes[0]
				Expect(process.DiskQuota).To(Equal("5MB"))
			})
		})
	})

	When("disk flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.Disk = "5MB"

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

		It("changes the disk of the app in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			Expect(workerProcess.DiskQuota).To(Equal(""))
			Expect(transformedManifest.Applications[0].DiskQuota).To(Equal("5MB"))
		})
	})

	When("disk flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.Disk = "5MB"

			originalManifest.Applications = []manifestparser.Application{
				{
					ApplicationModel: manifestparser.ApplicationModel{
						Processes: []manifestparser.ProcessModel{
							{Type: "worker", DiskQuota: "2MB"},
							{Type: "web", DiskQuota: "3MB"},
						},
						DiskQuota: "1MB",
					},
				},
			}
		})

		It("changes the disk of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			workerProcess := transformedManifest.Applications[0].Processes[0]
			webProcess := transformedManifest.Applications[0].Processes[1]
			Expect(workerProcess.DiskQuota).To(Equal("2MB"))
			Expect(webProcess.DiskQuota).To(Equal("5MB"))
			Expect(transformedManifest.Applications[0].DiskQuota).To(Equal("1MB"))

		})
	})

	When("disk flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.Disk = "5MB"

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
