package shovel

import (
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/entities"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/ccremer/clustercode-worker/process"
	log "github.com/sirupsen/logrus"
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
	log.Infof("Listening for task added events.")
	return instance
}

func (i *Instance) handleTaskAddedEvent(task *entities.TaskAddedEvent) {
	logEntry := log.WithField("task_id", task.JobID)
	logEntry.Info("Processing task.")
	p := process.New(task.Args, config.GetConfig())
	p.StartProcess()
	go p.PrintOutputLines()
	status := <-p.Cmd.Start()
	p.WaitForOutput()
	logEntry = logEntry.WithField("exit_code", status.Exit)
	if status.Error != nil || status.Exit > 0 {
		logEntry.Info("Task failed.")
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
