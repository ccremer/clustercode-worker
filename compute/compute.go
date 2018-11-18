package compute

import (
	"context"
	"errors"
	"fmt"
	"github.com/ccremer/clustercode-worker/api"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/ccremer/clustercode-worker/process"
	"github.com/go-cmd/cmd"
	"github.com/lestrrat-go/backoff"
	"os"
	"time"
)

var sliceCompleteChan chan messaging.SliceCompletedEvent
var taskCancelledChan chan messaging.TaskCancelledEvent

func Start() {

	sliceCompleteChan = make(chan messaging.SliceCompletedEvent)
	taskCancelledChan = make(chan messaging.TaskCancelledEvent)

	messaging.OpenSliceCompleteQueue(sliceCompleteChan)
	messaging.OpenSliceAddedQueue(handleSliceAddedEvent)
	log.Infof("Listening for task slices.")

	messaging.OpenTaskCancelledQueue(handleTaskCancelledEvent)
	log.Infof("Listening for task cancellations.")
}

const maxRetries = 9

var policy = backoff.NewExponential(
	backoff.WithMaxInterval(10*time.Minute),
	backoff.WithInterval(5*time.Second),
	backoff.WithMaxRetries(maxRetries),
)

func handleSliceAddedEvent(slice messaging.SliceAddedEvent) {

	b, cancel := policy.Start(context.Background())
	defer cancel()

	var count = 1
	for backoff.Continue(b) {
		log.Infof("Processing slice: %s, %d", slice.JobID, slice.SliceNr)
		ffmpeg := process.StartProcess(append(slice.Args, api.GetExtraArgsForProgress()[:]...))
		go process.PrintOutputLines(ffmpeg)
		go listenForCancelMessage(ffmpeg, &slice)
		err := waitForProcessToFinish(ffmpeg, &slice)
		api.ResetProgressMetrics()
		if err == nil {
			return
		} else {
			log.Warn(err.Error())
			log.Infof("Retrying soon, %d of %d", count, maxRetries+1)
			count++
		}
	}
	log.Errorf("Unfortunately, I cannot perform the task given. I have failed, I must go...")
	os.Exit(2)
}

func waitForProcessToFinish(ffmpeg *cmd.Cmd, slice *messaging.SliceAddedEvent) error {
	status := <-ffmpeg.Start()
	process.WaitForOutput(ffmpeg)
	if status.Error != nil || status.Exit > 0 {
		if status.Exit < 255 {
			slice.SetComplete(messaging.IncompleteAndRequeue)
			return errors.New(fmt.Sprintf("Task failed with exit code %d.", status.Exit))
		} else {
			slice.SetComplete(messaging.Complete)
			log.Infof("Task cancelled.")
			return nil
		}
	} else {
		slice.SetComplete(messaging.Complete)
		sendSliceCompletedMessage(slice)
		log.Infof("Task slice finished.")
		return nil
	}
}

func handleTaskCancelledEvent(event messaging.TaskCancelledEvent) {
	taskCancelledChan <- event
}

func listenForCancelMessage(
	ffmpeg *cmd.Cmd,
	currentTask *messaging.SliceAddedEvent,
) {
	for {
		event := <-taskCancelledChan
		if event.JobID == currentTask.JobID {
			log.Warnf("Cancelling task: %s", event.JobID)
			ffmpeg.Stop()
		} else {
			log.Debugf("TaskCancelledEvent %s does not match current job %s", event.JobID, currentTask.JobID)
		}
		event.SetComplete(messaging.Complete)
	}
}

func sendSliceCompletedMessage(slice *messaging.SliceAddedEvent) {
	sliceCompleteChan <- messaging.SliceCompletedEvent{
		SliceNr: slice.SliceNr,
		JobID:   slice.JobID,
	}
}
