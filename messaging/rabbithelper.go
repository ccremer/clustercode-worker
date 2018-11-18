package messaging

import (
	"github.com/ccremer/clustercode-worker/util"
	"github.com/streadway/amqp"
)

func failOnDeserialize(err error) {
	util.PanicOnErrorf("Could not deserialize message: %s. Please purge the invalid messages.", err)
}

func createChannel() *amqp.Channel {
	log.Debugf("Opening a new channel.")
	channel, err := connection.Channel()
	util.PanicOnErrorf("Failed to open channel: %[1]s", err)
	return channel
}

func createQueue(o *queueOptions, channel *amqp.Channel) amqp.Queue {
	log.Debugf("Creating queue %s", o.queueName)
	q, err := channel.QueueDeclare(
		o.queueName,
		o.durable,
		o.autoDelete,
		o.exclusive,
		o.noWait,
		o.args,
	)
	util.PanicOnErrorf("Failed to declare queue %[2]s: %[1]s", err, o.queueName)
	return q
}

func createExchange(o *queueOptions, channel *amqp.Channel) {
	log.Debugf("Creating exchange %s", o.exchangeName)
	err := channel.ExchangeDeclare(
		o.exchangeName,
		o.exchangeType,
		o.durable,
		o.autoDelete,
		o.internal,
		o.noWait,
		o.args)
	util.PanicOnErrorf("Failed to create exchange %[2]s: %[1]s", err, o.exchangeName)
}

func bindToExchange(o *queueOptions, channel *amqp.Channel) {
	log.Debugf("Binding queue %s to exchange %s", o.queueName, o.exchangeName)
	err := channel.QueueBind(
		o.queueName,
		o.routingKey,
		o.exchangeName,
		o.noWait,
		o.args)
	util.PanicOnErrorf("Failed to bind queue %[2]s: %[1]s", err, o.queueName)
}

func createConsumer(o *queueOptions, channel *amqp.Channel) <-chan amqp.Delivery {
	msgs, err := channel.Consume(
		o.queueName,
		o.consumerName,
		o.autoAck,
		o.exclusive,
		o.noLocal,
		o.noWait,
		o.args,
	)
	util.PanicOnErrorf("Failed to consume queue: %[2]s: %[1]s", err, o.queueName)
	return msgs
}

func ensureOnlyOneConsumerActive(channel *amqp.Channel) {
	prefetchCount, prefetchSize, global := 1, 0, false
	err := channel.Qos(
		prefetchCount,
		prefetchSize,
		global,
	)
	util.PanicOnErrorf("Failed to set QoS: %s", err)
}

func beginConsuming(msgs <-chan amqp.Delivery, callback messageReceivedCallback) {
	go func(msgs <-chan amqp.Delivery) {
		for msg := range msgs {
			log.Debugf("Received a message: %s", msg.Body)
			callback(&msg)
		}
	}(msgs)
}

func publish(options *queueOptions, channel *amqp.Channel, json string) {
	log.Debugf("Sending message: %s", json)
	channel.Publish(
		options.exchangeName,
		options.queueName,
		options.mandatory,
		options.immediate,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(json),
		})
}

func acknowledgeMessage(completionType CompletionType, delivery *amqp.Delivery) {
	switch completionType {
	case Complete:
		{
			delivery.Ack(false)
		}
	case Incomplete:
		{
			delivery.Nack(false, false)
		}
	case IncompleteAndRequeue:
		{
			delivery.Nack(false, true)
		}
	default:
		panic("This type is not expected here. This is a Programmer error!")
	}
}
