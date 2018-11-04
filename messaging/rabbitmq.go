package messaging

import (
    "github.com/ccremer/clustercode-worker/util"
    "github.com/micro/go-config"
    "github.com/streadway/amqp"
)

var connection *amqp.Connection

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

func OpenTaskAddedQueue(callback func(msg *Message)) {
    channel, q := createChannel(
        config.Get("rabbitmq", "channels", "task", "added").String("task-added"))

    ensureOnlyOneConsumerActive(channel)

    autoAck, exclusive, noLocal, noWait := false, false, false, false
    msgs, err := channel.Consume(
        q.Name,
        "",
        autoAck,
        exclusive,
        noLocal,
        noWait,
        nil,
    )
    util.PanicOnErrorf("Failed to register a consumer: %s", err)
    go func(callback func(message *Message)) {
        for msg := range msgs {
            log.Debugf("Received a message: %s", msg.Body)
            payload := Message{
                string(msg.Body),
                &msg,
            }
            callback(&payload)
        }
    }(callback)
}

func OpenTaskCompleteQueue(supplier chan Message) {
    channel, q := createChannel(
        config.Get("rabbitmq", "channels", "task", "completed").String("task-completed"))
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

func createChannel(queueName string) (*amqp.Channel, amqp.Queue) {
    log.Debugf("Opening a new channel %s...", queueName)
    channel, err := connection.Channel()
    util.PanicOnErrorf("Failed to open channel %[2]s: %[1]s", err, queueName)

    durable, autoDelete, exclusive, noWait := true, false, false, false
    q, err := channel.QueueDeclare(
        queueName,
        durable,
        autoDelete,
        exclusive,
        noWait,
        nil,
    )
    util.PanicOnErrorf("Failed to declare queue %[2]s: %[1]s", err, queueName)
    return channel, q
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
