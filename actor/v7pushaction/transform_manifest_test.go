package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/util/manifestparser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransformManifest", func() {
	var (
		pushActor *Actor
		baseManifest manifestparser.ParsedManifest
		flagOverrides FlagOverrides
		transformedManifest manifestparser.ParsedManifest
		executeErr error

		testFuncCallCount int
	)

	testTransformManifestFunc := func(manifest manifestparser.ParsedManifest, overrides FlagOverrides) (manifestparser.ParsedManifest, error) {
		testFuncCallCount += 1
		return manifest, nil
	}

	BeforeEach(func() {
		baseManifest = manifestparser.ParsedManifest{}
		flagOverrides = FlagOverrides{}
		testFuncCallCount = 0

		pushActor, _, _ = getTestPushActor()
		pushActor.TransformManifestSequence = []TransformManifestFunc{
			testTransformManifestFunc,
		}
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = pushActor.TransformManifest(baseManifest, flagOverrides)
	})

	It("calls each transform-manifest function", func() {
		Expect(testFuncCallCount).To(Equal(1))
		Expect(executeErr).NotTo(HaveOccurred())
		Expect(transformedManifest).To(Equal(baseManifest))
	})
})
