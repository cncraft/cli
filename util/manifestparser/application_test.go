package manifestparser_test

import (
	. "code.cloudfoundry.org/cli/util/manifestparser"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application", func() {
	Describe("Unmarshal", func() {
		var (
			rawYAML     []byte
			application Application
			executeErr  error
		)

		JustBeforeEach(func() {
			executeErr = yaml.Unmarshal(rawYAML, &application)
		})

		Context("when a name is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
name: spark
`)
			})

			It("unmarshals the name", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Name).To(Equal("spark"))
			})
		})

		Context("when a path is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
path: /my/path
`)
			})

			It("unmarshals the path", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Path).To(Equal("/my/path"))
			})
		})

		Context("when a docker map is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
docker:
  image: some-image
  username: some-username
`)
			})

			It("unmarshals the docker properties", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Docker.Image).To(Equal("some-image"))
				Expect(application.Docker.Username).To(Equal("some-username"))
			})
		})

		Context("when no-route is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
no-route: true
`)
			})

			It("unmarshals the no-route property", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.NoRoute).To(BeTrue())
			})
		})

		Context("when random-route is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
random-route: true
`)
			})

			It("unmarshals the random-route property", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.RandomRoute).To(BeTrue())
			})
		})

		Context("when buildpacks is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
buildpacks:
- ruby_buildpack
- java_buildpack
`)
			})

			It("unmarshals the buildpacks property", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Buildpacks).To(Equal([]string{"ruby_buildpack", "java_buildpack"}))
			})
		})

		Context("when stack is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
stack: cflinuxfs3
`)
			})

			It("unmarshals the stack property", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Stack).To(Equal("cflinuxfs3"))
			})
		})

		Context("when Processes are provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
processes: []
`)
			})

			It("unmarshals the processes property", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Processes).To(Equal([]ProcessModel{}))
			})
		})

		Context("process-level configuration", func() {
			Context("the Type command is always provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- type: web
`)
				})

				It("unmarshals the processes property with the type", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]ProcessModel{
						{Type: "web"},
					}))
				})
			})

			Context("when the start command is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- command: /bin/python
`)
				})

				It("unmarshals the processes property with the start command", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]ProcessModel{
						{StartCommand: "/bin/python"},
					}))
				})
			})

			Context("when a disk quota is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- disk_quota: 5GB
`)
				})

				It("unmarshals the processes property with the disk quota", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]ProcessModel{
						{DiskQuota: "5GB"},
					}))
				})
			})

			Context("when a health check endpoint is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- health-check-http-endpoint: https://localhost
`)
				})

				It("unmarshals the processes property with the health check endpoint", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]ProcessModel{
						{HealthCheckEndpoint: "https://localhost"},
					}))
				})
			})

			Context("when a health check type is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- health-check-type: http
`)
				})

				It("unmarshals the processes property with the health check type", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]ProcessModel{
						{HealthCheckType: "http"},
					}))
				})
			})

			Context("when a memory limit is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- memory: 512M
`)
				})

				It("unmarshals the processes property with the memory limit", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]ProcessModel{
						{Memory: "512M"},
					}))
				})
			})

			Context("when instances are provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- instances: 4
`)
				})

				It("unmarshals the processes property with instances", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]ProcessModel{
						{Instances: 4},
					}))
				})
			})
		})
	})
})
