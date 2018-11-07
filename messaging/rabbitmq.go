package messaging

import (
    "github.com/ccremer/clustercode-worker/util"
    "github.com/micro/go-config"
    "github.com/streadway/amqp"
)

var connection *amqp.Connection

type queueOptions struct {
    exclusive    bool
    durable      bool
    autoDelete   bool
    noWait       bool
    internal     bool
    routingKey   string
    exchangeName string
    exchangeType string
    queueName    string
    args         amqp.Table
    consumerName string
    autoAck      bool
    noLocal      bool
}

func newQueueOptions() queueOptions {
    return queueOptions{
        exclusive:    false,
        durable:      true,
        autoDelete:   false,
        noWait:       false,
        autoAck:      false,
        queueName:    "",
        args:         nil,
        routingKey:   "",
        exchangeName: "",
        exchangeType: "fanout",
        internal:     false,
        noLocal:      false,
        consumerName: "",
    }
}

func Connect() *amqp.Connection {
    if connection != nil {
        return connection
    }
    url := config.Get("rabbitmq", "url").String("amqp://guest:guest@rabbitmq:5672/")
    log.Infof("Connecting to %s", url)
    conn, err := amqp.Dial(url)
    util.PanicOnErrorf("A working connection to %[2]s is necessary: %[1]s", err, url)
    connection = conn
    return connection
}

func OpenSliceAddedQueue(callback func(msg *Message)) {
    options := newQueueOptions()
    options.queueName = config.Get("rabbitmq", "channels", "task", "added").String("task-added")
    channel := createChannel()
    q := createQueue(&options, channel)

    ensureOnlyOneConsumerActive(channel)

    options.consumerName = q.Name
    options.autoAck = false
    msgs := createConsumer(&options, channel)
    beginConsuming(msgs, callback)
}

func OpenSliceCompleteQueue(supplier chan Message) {
    options := newQueueOptions()
    options.queueName = config.Get("rabbitmq", "channels", "task", "completed").String("task-completed")
    channel := createChannel()
    q := createQueue(&options, channel)
    exchange, mandatory, immediate := "", false, false

    go func(channel *amqp.Channel) {
        for {
            msg := <-supplier
            channel.Publish(
                exchange,
                q.Name,
                mandatory,
                immediate,
                amqp.Publishing{
                    DeliveryMode: amqp.Persistent,
                    ContentType:  "application/json",
                    Body:         []byte(msg.Body),
                })
            log.Debugf("Sent message to queue %s: %s", q.Name, msg.Body)
        }
    }(channel)
}

func OpenTaskCancelledQueue(callback func(msg *Message)) {
    channel := createChannel()
    options := newQueueOptions()

    options.exchangeName = config.Get("rabbitmq", "channels", "task", "cancelled").String("task-cancelled")
    options.autoDelete = false
    options.durable = true
    createExchange(&options, channel)

    options.queueName = ""
    options.exclusive = true
    options.durable = false
    q := createQueue(&options, channel)

    options.queueName = q.Name
    bindToExchange(&options, channel)

    msgs := createConsumer(&options, channel)
    beginConsuming(msgs, callback)
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
    log.Debugf("Binding queue %s to exchangeName %s", o.queueName, o.exchangeName)
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

func beginConsuming(msgs <-chan amqp.Delivery, callback func(msg *Message)) {
    go func(msgs <-chan amqp.Delivery) {
        for msg := range msgs {
            log.Debugf("Received a message: %s", msg.Body)
            payload := Message{
                string(msg.Body),
                &msg,
            }
            callback(&payload)
        }
    }(msgs)
}
