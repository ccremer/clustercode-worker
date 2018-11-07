package compute

import (
    "github.com/ccremer/clustercode-worker/messaging"
    "github.com/ccremer/clustercode-worker/util"
    "github.com/go-cmd/cmd"
    "os/exec"
    "time"
)

func Start() {

    completeChan := make(chan messaging.Message)
    cancelChan := make(chan messaging.TaskCancelledEvent)

    messaging.OpenSliceAddedQueue(func(msg *messaging.Message) {
        task := messaging.SliceAddedEvent{}
        err := messaging.FromJson(msg.Body, &task)
        util.PanicOnErrorf("Could not deserialize message: %s. Please purge the invalid messages.", err)

        handleTaskAddedEvent(&task, cancelChan, msg, completeChan)
    })
    log.Infof("Listening for task slices.")

    messaging.OpenSliceCompleteQueue(completeChan)
    messaging.OpenTaskCancelledQueue(func(msg *messaging.Message) {
        event := messaging.TaskCancelledEvent{}
        err := messaging.FromJson(msg.Body, &event)
        util.PanicOnErrorf("Could not deserialize message: %s. Please purge the invalid messages.", err)

        cancelChan <- event

    })
}

func handleTaskAddedEvent(slice *messaging.SliceAddedEvent, cancelChan chan messaging.TaskCancelledEvent, msg *messaging.Message, completeChan chan messaging.Message) {
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
    go listenForCancelMessage(ffmpeg, cancelChan, slice)
    status := <-ffmpeg.Start()
    waitForOutput(ffmpeg)
    log.Debugf("Process finished with exit code %d.", status.Exit)
    if status.Error != nil || status.Exit > 1 {
        msg.SetIncompleteAndRequeue()
    } else {
        msg.SetComplete()
        sendCompletedMessage(*slice, completeChan)
    }
    log.Infof("Task slice finished.")
}

func listenForCancelMessage(
    ffmpeg *cmd.Cmd,
    cancelChan chan messaging.TaskCancelledEvent,
    currentTask *messaging.SliceAddedEvent,
) {
    for {
        event := <-cancelChan
        if event.JobID == currentTask.JobID {
            log.Warnf("Cancelling task: %s", event.JobID)
            ffmpeg.Stop()
            // TODO: Ack message, else it gets re-scheduled!
        } else {
            log.Debugf("TaskCancelledEvent %s does not match current job %s", event.JobID, currentTask.JobID)
        }
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

func sendCompletedMessage(slice messaging.SliceAddedEvent, completeChan chan messaging.Message) {
    json, _ := messaging.ToJson(messaging.SliceCompleteEvent{
        SliceNr: slice.SliceNr,
        JobID:   slice.JobID,
    })
    completeChan <- messaging.Message{
        Body: json,
    }
}

func waitForOutput(ffmpeg *cmd.Cmd) {
    for len(ffmpeg.Stdout) > 0 || len(ffmpeg.Stderr) > 0 {
        time.Sleep(10 * time.Millisecond)
    }
}
