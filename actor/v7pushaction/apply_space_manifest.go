package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	log "github.com/sirupsen/logrus"
)

func (actor Actor) ApplySpaceManifest(pushPlans []PushPlan, rawManifest []byte) <-chan *PushEvent {
	pushEventStream := make(chan *PushEvent)

	go func() {
		log.Debug("starting apply manifest go routine")
		defer close(pushEventStream)
		//
		var warnings v7action.Warnings
		var err error
		var successEvent Event

		pushEventStream <- &PushEvent{Event: ApplyManifest}
		warnings, err = actor.V7Actor.SetSpaceManifest(pushPlans[0].SpaceGUID, rawManifest, pushPlans[0].NoRouteFlag)
		successEvent = ApplyManifestComplete

		if err != nil {
			pushEventStream <- &PushEvent{Err: err, Warnings: Warnings(warnings), Plan: pushPlans[0]}
			return
		}

		pushEventStream <- &PushEvent{Event: successEvent, Err: nil, Warnings: Warnings(warnings), Plan: pushPlans[0]}
	}()

	return pushEventStream
}

