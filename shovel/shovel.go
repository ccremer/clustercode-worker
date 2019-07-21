package shovel

import (
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/entities"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/ccremer/clustercode-worker/process"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strconv"
	"time"
)

type (
	Instance struct {
		MessagingService   *messaging.RabbitMqService
		config             config.ConfigMap
		taskCompletedQueue *messaging.ChannelConfig
	}
)

func NewInstance(service *messaging.RabbitMqService) Instance {
	instance := Instance{
		MessagingService:   service,
		taskCompletedQueue: getTaskCompletedQueue(),
	}
	service.AddChannelConfig(instance.taskCompletedQueue)
	service.AddChannelConfig(getTaskAddedQueue(instance.handleTaskAddedEvent))
	return instance
}

func (i *Instance) handleTaskAddedEvent(task *entities.TaskAddedEvent) {
	logEntry := log.WithField("task_id", task.JobID)
	logEntry.Info("Processing task.")
	var args []string
	if task.Type == "SPLIT" {
		args = config.GetConfig().Api.Ffmpeg.SplitArgs
	} else if task.Type == "MERGE" {
		args = config.GetConfig().Api.Ffmpeg.MergeArgs
	}
	c := config.GetConfig()
	p := process.New(args, config.GetConfig())

	switch c.Api.Ffmpeg.Debug {
	case config.ApiFfmpegDebugLog:
		if log.IsLevelEnabled(log.DebugLevel) {
			p.SetBothOutHandlers(process.PrintToDebugHandler)
		}
	case config.ApiFfmpegDebugOut:
		p.StdOutHandler = process.PrintToStdOutHandler
		p.StdErrHandler = process.PrintToStdErrHandler
	}

	status := <-p.StartProcess(map[string]string{
		"${INPUT}":      task.Media.GetSubstitutedPath(c.Input.Dir),
		"${OUTPUT}":     c.Output.Dir,
		"${TMP}":        c.Output.TmpDir,
		"${JOB}":        task.JobID,
		"${FORMAT}": 	 filepath.Ext(task.Media.RelativePath()),
		"${SLICE_SIZE}": strconv.Itoa(task.SliceSize),
	})

	logEntry = logEntry.WithField("exit_code", status)

	if status.Error != nil || status.Exit > 0 {
		logEntry.Info("Task failed.")
		time.Sleep(30 * time.Second)
		task.SetComplete(entities.IncompleteAndRequeue)
	} else {
		logEntry.Info("Task finished.")
		task.SetComplete(entities.Complete)
		i.sendTaskCompletedMessage(task)
	}
}

func (i *Instance) sendTaskCompletedMessage(task *entities.TaskAddedEvent) {
	payload, err := entities.ToJson(entities.TaskCompletedEvent{
		JobID: task.JobID,
	})
	if err == nil {
		i.MessagingService.Publish(i.taskCompletedQueue, payload)
		task.SetComplete(entities.Complete)
	} else {
		task.SetComplete(entities.IncompleteAndRequeue)
		log.WithError(err).WithField("task_id", task.JobID).Error("Could not serialize to XML. Requeueing.")
	}
}
