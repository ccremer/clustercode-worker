package shovel

import (
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/entities"
	"github.com/ccremer/clustercode-worker/messaging"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func createTaskAddedQueue(consumer func(task *entities.TaskAddedEvent) TaskResult) *messaging.ChannelConfig {
	cfg := config.ConvertToChannelConfig(config.GetConfig().RabbitMq.Channels.Task.Added)
	cfg.Consumer = func(d *amqp.Delivery) {
		if event, err := entities.DeserializeTaskAddedEvent(d); err == nil {
			consumer(event)
		} else {
			log.WithError(err).Fatal("Could not deserialize message.")
		}
	}
	return cfg
}

func getTaskCompletedQueue() *messaging.ChannelConfig {
	return config.ConvertToChannelConfig(config.GetConfig().RabbitMq.Channels.Task.Completed)
}

func (i *Instance) sendTaskCompletedMessage(result *TaskResult) {
	task := result.TaskAddedEvent
	payload, err := entities.ToJson(entities.TaskCompletedEvent{
		JobID:  task.JobID,
		Amount: result.SliceCountResult,
		Type:   result.TaskType,
	})
	if err == nil {
		log.Debug(i.taskCompletedQueue.ExchangeOptions.ExchangeName)
		i.MessagingService.Publish(i.taskCompletedQueue, payload)
	} else {
		log.WithError(err).WithField("task_id", task.JobID).Panic("Could not serialize to JSON.")
	}
}
