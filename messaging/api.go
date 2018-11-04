package messaging

import (
    json2 "encoding/json"
    "github.com/streadway/amqp"
)

type PreTask struct {
    JobID     string `json:"job_id"`
    File      string
    Priority  int
    SliceSize int    `json:"slice_size"`
    FileHash  string `json:"file_hash"`
    Args      []string
}

type TaskAddedEvent struct {
    JobID string `json:"job_id"`
    Slice int    `json:"slice"`
    Args      []string
}

type TaskCompleteEvent struct {
    JobID    string `json:"job_id"`
    FileHash string `json:"file_hash"`
    SliceNr  int    `json:"slice_nr"`
}

type TaskCancelledEvent struct {
    JobID string `json:"job_id"`
}

type Message struct {
    Body     string
    delivery *amqp.Delivery
}

func (msg *Message) SetComplete() {
    msg.delivery.Ack(false)
}

func (msg *Message) SetIncomplete() {
    msg.delivery.Nack(false, false)
}

func (msg *Message) SetIncompleteAndRequeue() {
    msg.delivery.Nack(false, true)
}

func FromJson(json string, value interface{}) error {
    arr := []byte(json)
    err := json2.Unmarshal(arr, &value)
    return err
}

func ToJson(value interface{}) (string, error) {
    json, err := json2.Marshal(&value)
    if err == nil {
        return string(json[:]), nil
    } else {
        return "", err
    }
}
