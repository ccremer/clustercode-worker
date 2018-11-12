package compute

import (
    "github.com/ccremer/clustercode-worker/messaging"
    "github.com/ccremer/clustercode-worker/util"
    "github.com/go-cmd/cmd"
    "os/exec"
    "time"
)

var taskCompleteChan chan messaging.SliceCompleteEvent
var taskCancelledChan chan messaging.TaskCancelledEvent

func Start() {

    taskCompleteChan = make(chan messaging.SliceCompleteEvent)
    taskCancelledChan = make(chan messaging.TaskCancelledEvent)

    messaging.OpenSliceAddedQueue(handleTaskAddedEvent)
    log.Infof("Listening for task slices.")

    messaging.OpenSliceCompleteQueue(taskCompleteChan)
    messaging.OpenTaskCancelledQueue(handleTaskCancelledEvent)
    log.Infof("Listening for task cancellations.")
}

func handleTaskAddedEvent(slice messaging.SliceAddedEvent) {
    log.Infof("Processing slice: %s", slice.JobID)
    path, err := exec.LookPath("ffmpeg")
    util.PanicOnError(err)
    //command := path + " -hide_banner -y -t 20 -s 640x480 -f rawvideo -pix_fmt rgb24 -r 25 -i /dev/zero vendor/empty.mpeg"
    cmdOptions := cmd.Options{
        Buffered:  false,
        Streaming: true,
    }
    ffmpeg := cmd.NewCmdOptions(cmdOptions, path, slice.Args[:]...)
    log.Infof("Starting process: %s", ffmpeg.Args)
    ffmpeg.Start()
    go printOutputLines(ffmpeg)
    go listenForCancelMessage(ffmpeg, &slice)
    status := <-ffmpeg.Start()
    waitForOutput(ffmpeg)
    log.Debugf("Process finished with exit code %d.", status.Exit)
    if status.Error != nil || status.Exit > 1 {
        if status.Exit < 255 {
            slice.SetComplete(messaging.IncompleteAndRequeue)
            log.Infof("Task failed.")
        } else {
            slice.SetComplete(messaging.Complete)
            log.Infof("Task cancelled.")
        }
    } else {
        slice.SetComplete(messaging.Complete)
        sendCompletedMessage(&slice)
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

func printOutputLines(cmd *cmd.Cmd) {
    for {
        select {
        case line := <-cmd.Stdout:
            log.Info(line)
        case line := <-cmd.Stderr:
            log.Error(line)
        }
    }
}

func sendCompletedMessage(slice *messaging.SliceAddedEvent) {
    taskCompleteChan <- messaging.SliceCompleteEvent{
        SliceNr: slice.SliceNr,
        JobID:   slice.JobID,
    }
}

func waitForOutput(ffmpeg *cmd.Cmd) {
    for len(ffmpeg.Stdout) > 0 || len(ffmpeg.Stderr) > 0 {
        time.Sleep(10 * time.Millisecond)
    }
}
