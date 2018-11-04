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

    messaging.OpenTaskAddedQueue(func(msg *messaging.Message) {
        task := messaging.TaskAddedEvent{}
        err := messaging.FromJson(string(msg.Body), &task)
        util.PanicOnErrorf("Could not deserialize message: %s. Please purge the invalid messages.", err)

        log.Infof("Processing slice: %s", task.JobID)
        path, err := exec.LookPath("ffmpeg")
        util.PanicOnError(err)
        //command := path + " -hide_banner -y -t 20 -s 640x480 -f rawvideo -pix_fmt rgb24 -r 25 -i /dev/zero vendor/empty.mpeg"

        cmdOptions := cmd.Options{
            Buffered:  false,
            Streaming: true,
        }

        ffmpeg := cmd.NewCmdOptions(cmdOptions, path, task.Args[:]...)

        log.Infof("Starting process: %s", ffmpeg.Args)
        ffmpeg.Start()

        go func(cmd *cmd.Cmd) {
            for {
                select {
                case line := <-cmd.Stdout:
                    log.Info(line)
                case line := <-cmd.Stderr:
                    log.Error(line)
                }
            }
        }(ffmpeg)

        status := <-ffmpeg.Start()
        waitForOutput(ffmpeg)
        log.Debugf("Process finished with exit code %d", status.Exit)

        if status.Error != nil || status.Exit > 1 {
            msg.SetIncompleteAndRequeue()
        } else {
            msg.SetComplete()
            sendCompletedMessage(task, completeChan)
        }

        log.Infof("Task slice finished")
    })
    log.Infof("Listening for task slices.")

    messaging.OpenTaskCompleteQueue(completeChan)
}

func sendCompletedMessage(task messaging.TaskAddedEvent, completeChan chan messaging.Message) {
    json, _ := messaging.ToJson(messaging.TaskCompleteEvent{
        SliceNr: task.Slice,
        JobID:   task.JobID,
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

