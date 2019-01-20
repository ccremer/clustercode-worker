package shovel

import (
	"github.com/ccremer/clustercode-api-gateway/entities"
	"github.com/ccremer/clustercode-worker/process"
	log "github.com/sirupsen/logrus"
)

var taskCompleteChan chan entities.TaskCompletedEvent

func Start() {
	taskCompleteChan = make(chan entities.TaskCompletedEvent)

	//entities.OpenTaskCompleteQueue(taskCompleteChan)
	//entities.OpenTaskAddedQueue(handleTaskAddedEvent)
	log.Infof("Listing for task added events.")
}

func handleTaskAddedEvent(task entities.TaskAddedEvent) {
	log.Infof("Processing task: %s", task.JobID)
	ffmpeg := process.StartProcess(task.Args)
	go process.PrintOutputLines(ffmpeg)
	status := <-ffmpeg.Start()
	process.WaitForOutput(ffmpeg)
	log.Debugf("Process finished with exit code %d.", status.Exit)
	if status.Error != nil || status.Exit > 0 {
		log.Infof("Task failed.")
		task.SetComplete(entities.IncompleteAndRequeue)
	} else {
		log.Infof("Task finished.")
		task.SetComplete(entities.Complete)
		sendTaskCompletedMessage(&task)
	}
}

func sendTaskCompletedMessage(slice *entities.TaskAddedEvent) {
	taskCompleteChan <- entities.TaskCompletedEvent{
		JobID: slice.JobID,
	}
}
