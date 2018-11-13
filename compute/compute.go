package compute

import (
    "github.com/ccremer/clustercode-worker/messaging"
    "github.com/ccremer/clustercode-worker/process"
    "github.com/go-cmd/cmd"
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

func handleSliceAddedEvent(slice messaging.SliceAddedEvent) {
    log.Infof("Processing slice: %s, %d", slice.JobID, slice.SliceNr)
    ffmpeg := process.StartProcess(slice.Args)
    go process.PrintOutputLines(ffmpeg)
    go listenForCancelMessage(ffmpeg, &slice)
    waitForProcessToFinish(ffmpeg, slice)
}

func waitForProcessToFinish(ffmpeg *cmd.Cmd, slice messaging.SliceAddedEvent) {
    status := <-ffmpeg.Start()
    process.WaitForOutput(ffmpeg)
    log.Debugf("Process finished with exit code %d.", status.Exit)
    if status.Error != nil || status.Exit > 0 {
        if status.Exit < 255 {
            slice.SetComplete(messaging.IncompleteAndRequeue)
            log.Infof("Task failed.")
        } else {
            slice.SetComplete(messaging.Complete)
            log.Infof("Task cancelled.")
        }
    } else {
        slice.SetComplete(messaging.Complete)
        sendSliceCompletedMessage(&slice)
        log.Infof("Task slice finished.")
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
