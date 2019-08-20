package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ApplySpaceManifest", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		pushPlans           []PushPlan

		spaceGUID string

		eventStream <-chan *PushEvent
		appName1 = "app-name1"

		rawManifest []byte
	)

	BeforeEach(func() {
		actor, fakeV7Actor, _ = getTestPushActor()

		spaceGUID = "space"

		rawManifest = []byte("some manifest")
	})

	AfterEach(func() {
		Eventually(streamsDrainedAndClosed(eventStream)).Should(BeTrue())
	})

	JustBeforeEach(func() {
		eventStream = actor.ApplySpaceManifest(pushPlans, rawManifest)
	})

	When("applying the manifest fails", func() {
		BeforeEach(func() {
			pushPlans = []PushPlan{{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}}}
			fakeV7Actor.SetSpaceManifestReturns(v7action.Warnings{"apply-manifest-warnings"}, errors.New("some-error"))
		})

		It("returns the error and exits", func() {
			Eventually(eventStream).Should(Receive(Equal(&PushEvent{Event: ApplyManifest})))
			Eventually(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(1))
			actualSpaceGuid, actualManifest, _ := fakeV7Actor.SetSpaceManifestArgsForCall(0)
			Expect(actualSpaceGuid).To(Equal(spaceGUID))
			Expect(actualManifest).To(Equal(rawManifest))

			Eventually(eventStream).Should(Receive(Equal(&PushEvent{
				Warnings: Warnings{"apply-manifest-warnings"},
				Plan:     PushPlan{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}},
				Err:      errors.New("some-error"),
			})))
		})
	})

	When("There is a single pushPlan", func() {

		BeforeEach(func() {
			pushPlans = []PushPlan{{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}, NoRouteFlag: true}}
			fakeV7Actor.SetSpaceManifestReturns(v7action.Warnings{"apply-manifest-warnings"}, nil)
		})

		It("applies the app specific manifest", func() {
			Eventually(eventStream).Should(Receive(Equal(&PushEvent{Event: ApplyManifest})))
			Eventually(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(1))
			actualSpaceGUID, actualManifest, actualNoRoute := fakeV7Actor.SetSpaceManifestArgsForCall(0)
			Expect(actualManifest).To(Equal(rawManifest))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
			Expect(actualNoRoute).To(BeTrue())

			Eventually(eventStream).Should(Receive(Equal(&PushEvent{
				Event:    ApplyManifestComplete,
				Warnings: Warnings{"apply-manifest-warnings"},
				Plan:     PushPlan{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}, NoRouteFlag: true},
			})))
		})
	})
})
