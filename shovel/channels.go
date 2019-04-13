package shovel

import (
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/entities"
	"github.com/ccremer/clustercode-worker/messaging"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func getTaskAddedQueue(consumer func(task *entities.TaskAddedEvent)) *messaging.ChannelConfig {
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
