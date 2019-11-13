package compute

import (
	"github.com/ccremer/clustercode-worker/entities"
	"github.com/ccremer/clustercode-worker/messaging"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func NewComputeInstance(service *messaging.RabbitMqService) Instance {
	i := Instance{
		MessagingService:  service,
		sliceCompleteChan: make(chan *entities.SliceCompletedEvent),
	}
	service.AddChannelConfig(createSliceCompleteQueue())
	service.AddChannelConfig(createSliceAddedQueue(i.handleSliceAddedEvent))
	service.AddChannelConfig(createTaskCancelledQueue(i.handleTaskCancelledEvent))

	return i
}

func (i *Instance) handleTaskCancelledEvent(d *amqp.Delivery, event *entities.TaskCancelledEvent) {
	i.CancelTask(event)
	messaging.AcknowledgeMessage(d, messaging.Complete)
}

func (i *Instance) handleSliceAddedEvent(d *amqp.Delivery, sliceAddedEvent *entities.SliceAddedEvent) {
	result := i.ProcessSlice(sliceAddedEvent)
	logEvent := log.WithFields(log.Fields{
		"job_id":    result.SliceAddedEvent.JobID,
		"cancelled": result.Cancelled,
	})
	if result.Cancelled {
		logEvent.Warn("Task cancelled.")
		messaging.AcknowledgeMessage(d, messaging.Incomplete)
	} else if result.Error != nil {
		logEvent.
			WithError(result.Error).
			Error("Task failed.")
		messaging.AcknowledgeMessage(d, messaging.IncompleteAndRequeue)
	} else {
		messaging.AcknowledgeMessage(d, messaging.Complete)
		logEvent.Info("Task completed.")
	}
}
