package messaging

import (
	json2 "encoding/json"
	"github.com/streadway/amqp"
)

const (
	Complete             CompletionType = 0
	Incomplete           CompletionType = 1
	IncompleteAndRequeue CompletionType = 2
)

type (
	CompletionType int
	TaskAddedEvent struct {
		JobID     string `json:"job_id"`
		File      string
		Priority  int
		SliceSize int    `json:"slice_size"`
		FileHash  string `json:"file_hash"`
		Args      []string
		delivery  *amqp.Delivery
	}
	TaskCompletedEvent struct {
		JobID string `json:"job_id"`
	}
	TaskCancelledEvent struct {
		JobID    string `json:"job_id"`
		delivery *amqp.Delivery
	}
	SliceAddedEvent struct {
		JobID    string `json:"job_id"`
		SliceNr  int    `json:"slice_nr"`
		Args     []string
		delivery *amqp.Delivery
	}
	SliceCompletedEvent struct {
		JobID    string `json:"job_id"`
		FileHash string `json:"file_hash"`
		SliceNr  int    `json:"slice_nr"`
	}
)

func fromJson(json string, value interface{}) error {
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
