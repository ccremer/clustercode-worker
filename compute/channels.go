package compute

import (
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/entities"
	"github.com/ccremer/clustercode-worker/messaging"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func createSliceCompleteQueue() *messaging.ChannelConfig {
	return config.ConvertToChannelConfig(config.GetConfig().RabbitMq.Channels.Slice.Completed)
}

func createSliceAddedQueue(consumer func(d *amqp.Delivery, slice *entities.SliceAddedEvent)) *messaging.ChannelConfig {
	cfg := config.ConvertToChannelConfig(config.GetConfig().RabbitMq.Channels.Slice.Added)
	cfg.Consumer = func(d *amqp.Delivery) {
		if event, err := entities.DeserializeSliceAddedEvent(d); err == nil {
			consumer(d, event)
		} else {
			log.WithError(err).Fatal("Could not deserialize message.")
		}
	}
	return cfg
}

func createTaskCancelledQueue(consumer func(d *amqp.Delivery, event *entities.TaskCancelledEvent)) *messaging.ChannelConfig {
	cfg := config.ConvertToChannelConfig(config.GetConfig().RabbitMq.Channels.Task.Cancelled)
	cfg.Consumer = func(d *amqp.Delivery) {
		if event, err := entities.DeserializeTaskCancelledEvent(d); err == nil {
			consumer(d, event)
		} else {
			log.WithError(err).Fatal("Could not deserialize message.")
		}
	}
	return cfg
}
