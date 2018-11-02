package messages

import (
    json2 "encoding/json"
    "github.com/ccremer/clustercode-worker/util"
    "github.com/streadway/amqp"
    "os"
)

type PreTask struct {
    JobId     string
    File      string
    Priority  int
    SliceSize int    `json:"slice_size"`
    FileHash  string `json:"file_hash"`
    Args      []string
}

func (value *PreTask) FromJson(json string) error {
    arr := []byte(json)
    err := json2.Unmarshal(arr, value)
    return err
}

func (value *PreTask) ToJson() (string, error) {
    json, err := json2.Marshal(&value)
    if err == nil {
        return string(json[:]), nil
    } else {
        return "", err
    }
}

var connection *amqp.Connection

func Connect() {
    url := os.Getenv("CC_RABBITMQ_URL")
    conn, err := amqp.Dial(url)
    util.PanicOnErrorf("A working connection to rabbitmq is necessary: %ss", err)
    connection = conn
}

func OpenTaskQueue() (<-chan *amqp.Delivery) {
    ch, err := connection.Channel()
    util.PanicOnErrorf("Failed to open a channel: %s", err)

    durable, autoDelete, exclusive, noWait := true, false, false, false
    q, err := ch.QueueDeclare(
        "task_queue",
        durable,
        autoDelete,
        exclusive,
        noWait,
        nil,
    )
    util.PanicOnErrorf("Failed to declare a queue: %s", err)

    // We can only consume 1 at a time
    prefetchCount, prefetchSize, global := 1, 0, false
    err = ch.Qos(
        prefetchCount,
        prefetchSize,
        global,
    )
    util.PanicOnErrorf("Failed to set QoS: %s", err)

    autoAck, exclusive, noLocal, noWait := false, false, false, false
    msgs, err := ch.Consume(
        q.Name,
        "",
        autoAck,
        exclusive,
        noLocal,
        noWait,
        nil,
    )
    util.PanicOnErrorf("Failed to register a consumer: %s", err)

    callback := make(chan *amqp.Delivery)
    go func() {
        for d := range msgs {
            log.Debugf("Received a message: %s", d.Body)
            callback <- &d
            log.Debugf("Done")
        }
    }()

    return callback
}
