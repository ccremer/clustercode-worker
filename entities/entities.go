package entities

import (
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/messaging"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"net/url"
	"path/filepath"
	"strconv"
	"time"
)

const (
	Complete             CompletionType = 0
	Incomplete           CompletionType = 1
	IncompleteAndRequeue CompletionType = 2
	StdInFileDescriptor                 = 0
	StdOutFileDescriptor                = 1
	StdErrFileDescriptor                = 2
	Merge                TaskType       = "MERGE"
	Split                TaskType       = "SPLIT"
)

type (
	CompletionType int
	TaskType string
	TaskAddedEvent struct {
		JobID     string `json:"job_id"`
		Media     Media
		SliceSize int      `json:"slice_size"`
		Type      TaskType `json:"type"`
		delivery  *amqp.Delivery
	}
	TaskCompletedEvent struct {
		JobID  string `json:"job_id"`
		Amount int
		Type   TaskType
	}
	TaskCancelledEvent struct {
		JobID    string `json:"job_id"`
		delivery *amqp.Delivery
	}
	SliceAddedEvent struct {
		JobID    string `json:"job_id"`
		SliceNr  int
		Args     []string `xml:"Args>Arg,omitempty"`
		delivery *amqp.Delivery
	}
	SliceCompletedEvent struct {
		JobID      string `json:"job_id"`
		FileHash   string `xml:",omitempty"`
		SliceNr    int
		StdStreams []StdStream `xml:"StdStreams>L,omitempty"`
	}
	StdStream struct {
		FD   int    `xml:"fd,attr"`
		Line string `xml:",innerxml"`
	}
	Media struct {
		FileHash string
		Path     *url.URL
	}
	Message interface {
		SetComplete(completionType CompletionType)
	}
)

func DeserializeSliceAddedEvent(d *amqp.Delivery) (*SliceAddedEvent, error) {
	event := &SliceAddedEvent{
		delivery: d,
	}
	if err := FromJson(string(d.Body), event); err != nil {
		return nil, err
	}
	return event, nil
}

func DeserializeTaskCancelledEvent(d *amqp.Delivery) (*TaskCancelledEvent, error) {
	event := &TaskCancelledEvent{
		delivery: d,
	}
	if err := FromJson(string(d.Body), event); err != nil {
		return nil, err
	}
	return event, nil
}

func DeserializeTaskAddedEvent(d *amqp.Delivery) (*TaskAddedEvent, error) {
	event := &TaskAddedEvent{
		delivery: d,
	}
	if err := FromJson(string(d.Body), event); err != nil {
		return nil, err
	}
	return event, nil
}

func (e TaskCancelledEvent) SetComplete(completionType CompletionType) {
	acknowledgeMessage(completionType, e.delivery)
}

func (e SliceAddedEvent) SetComplete(completionType CompletionType) {
	acknowledgeMessage(completionType, e.delivery)
}

func (e TaskAddedEvent) SetComplete(completionType CompletionType) {
	acknowledgeMessage(completionType, e.delivery)
}

func (m Media) Priority() int {
	port, err := strconv.Atoi(m.Path.Port())
	if err == nil {
		return port
	} else {
		return 0
	}
}

func (m Media) RelativePath() string {
	if m.Path == nil {
		return ""
	}
	return m.Path.RequestURI()
}

func (m Media) GetSubstitutedPath(basePath string) string {
	if m.Path == nil {
		return ""
	}
	u := m.Path
	path, err := url.PathUnescape(u.RequestURI())
	if err != nil {
		log.WithField("uri", u.RequestURI()).Warn("Cannot parse URI, trying raw.")
		path = u.RequestURI()
	}
	return filepath.Join(basePath, u.Port(), path)
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
		log.WithField("type", completionType).Panic("type is not expected here")
	}
}

func Initialize() {
	log.Info("Called initialize")

	cfg := config.GetConfig()
	service := messaging.NewRabbitMqService(cfg.RabbitMq.Url)

	taskCancelledConfig := &messaging.ChannelConfig{
		ExchangeOptions: &cfg.RabbitMq.Channels.Task.Cancelled.Exchange,
		QueueOptions:    &cfg.RabbitMq.Channels.Task.Cancelled.Queue,
		Consumer: func(d *amqp.Delivery) {
			event := TaskCancelledEvent{}
			err := FromJson(string(d.Body), &event)
			failOnDeserialize(err, d.Body)
			event.delivery = d
			log.Info(event)
		}}

	log.Debug(taskCancelledConfig)
	service.Start(taskCancelledConfig)
	//service.AddChannelConfig(taskCancelledConfig)

	go func() {
		for i := 0; true; i++ {
			service.Publish(taskCancelledConfig, "a"+strconv.Itoa(i))
			time.Sleep(10 * time.Second)
		}
	}()
	go func() {
		for i := 0; true; i++ {
			service.Publish(taskCancelledConfig, "b"+strconv.Itoa(i))
			time.Sleep(12 * time.Second)
		}
	}()
}
