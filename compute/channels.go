package compute

import (
	"github.com/ccremer/clustercode-api-gateway/entities"
	"github.com/ccremer/clustercode-api-gateway/messaging"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

func openSliceCompleteQueue() *messaging.ChannelConfig {

	qOpts := messaging.NewQueueOptions()

	readQueueOptions("rabbitmq.channels.slice.completed.queue", qOpts)
	return &messaging.ChannelConfig{
		QueueOptions: qOpts,
		//ExchangeOptions: messaging.NewExchangeOptions(),
	}
}

func openSliceAddedQueue(consumer func(slice *entities.SliceAddedEvent)) *messaging.ChannelConfig {

	qOpts := messaging.NewQueueOptions()

	readQueueOptions("rabbitmq.channels.slice.added.queue", qOpts)

	return &messaging.ChannelConfig{
		QueueOptions: qOpts,
		Consumer: func(d *amqp.Delivery) {
			if event, err := entities.DeserializeSliceAddedEvent(d); err == nil {
				consumer(event)
			} else {
				log.WithField("error", err).Fatal("could not deserialize message")
			}
		},
	}
}

func openTaskCancelledQueue(consumer func(event *entities.TaskCancelledEvent)) *messaging.ChannelConfig {
	eOpts := messaging.NewExchangeOptions()
	qOpts := messaging.NewQueueOptions()
	readExchangeOptions("rabbitmq.channels.task.cancelled.exchange", eOpts)
	readQueueOptions("rabbitmq.channels.task.cancelled.queue", qOpts)
	return &messaging.ChannelConfig{
		ExchangeOptions: eOpts,
		QueueOptions:qOpts,
		Consumer: func(d *amqp.Delivery) {
			if event, err := entities.DeserializeTaskCancelledEvent(d); err == nil {
				consumer(event)
			} else {
				log.WithField("error", err).Fatal("could not deserialize message")
			}
		},
	}
}

func readQueueOptions(key string, opts interface{}) {
	if err := viper.UnmarshalKey(key, opts); err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err,
		}).Fatal("could not read config value")
	}
}

func readExchangeOptions(key string, opts interface{}) {
	if err := viper.UnmarshalKey(key, opts); err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err,
		}).Fatal("could not read config value")
	}
}
