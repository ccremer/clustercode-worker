package shovel

import (
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/ccremer/clustercode-worker/process"
)

var taskCompleteChan chan messaging.TaskCompletedEvent

func Start() {
	taskCompleteChan = make(chan messaging.TaskCompletedEvent)

	messaging.OpenTaskCompleteQueue(taskCompleteChan)
	messaging.OpenTaskAddedQueue(handleTaskAddedEvent)
	log.Infof("Listing for task added events.")
}

func handleTaskAddedEvent(task messaging.TaskAddedEvent) {
	log.Infof("Processing task: %s", task.JobID)
	ffmpeg := process.StartProcess(task.Args)
	go process.PrintOutputLines(ffmpeg)
	status := <-ffmpeg.Start()
	process.WaitForOutput(ffmpeg)
	log.Debugf("Process finished with exit code %d.", status.Exit)
	if status.Error != nil || status.Exit > 0 {
		log.Infof("Task failed.")
		task.SetComplete(messaging.IncompleteAndRequeue)
	} else {
		log.Infof("Task finished.")
		task.SetComplete(messaging.Complete)
		sendTaskCompletedMessage(&task)
	}
}

func sendTaskCompletedMessage(slice *messaging.TaskAddedEvent) {
	taskCompleteChan <- messaging.TaskCompletedEvent{
		JobID: slice.JobID,
	}
}
