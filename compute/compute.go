package compute

import (
	"errors"
	"fmt"
	"github.com/ccremer/clustercode-api-gateway/entities"
	"github.com/ccremer/clustercode-api-gateway/messaging"
	"github.com/ccremer/clustercode-worker/api"
	"github.com/ccremer/clustercode-worker/process"
	"github.com/efritz/backoff"
	"github.com/go-cmd/cmd"
	log "github.com/sirupsen/logrus"
	"time"
)

var sliceCompleteChan = make(chan *entities.SliceCompletedEvent)
var taskCancelledChan = make(chan *entities.TaskCancelledEvent)

func Start(s *messaging.RabbitMqService) {
	s.AddChannelConfig(openSliceCompleteQueue())

	s.AddChannelConfig(openSliceAddedQueue(handleSliceAddedEvent))
	log.Infof("Listening for task slices.")

	s.AddChannelConfig(openTaskCancelledQueue(handleTaskCancelledEvent))
	log.Infof("Listening for task cancellations.")

	service = s
}

var service *messaging.RabbitMqService
var sliceCompleteChannel *messaging.ChannelConfig

var b = backoff.NewConstantBackoff(10 * time.Second)

func handleSliceAddedEvent(sliceAddedEvent *entities.SliceAddedEvent) {
	for {
		log.WithFields(log.Fields{
			"job_id":   sliceAddedEvent.JobID,
			"slice_nr": sliceAddedEvent.SliceNr,
		}).Infof("Processing SliceAddedEvent")
		ffmpeg := process.StartProcess(append(sliceAddedEvent.Args, api.GetExtraArgsForProgress()[:]...))
		sliceCompletedEvent := &entities.SliceCompletedEvent{
			JobID:   sliceAddedEvent.JobID,
			SliceNr: sliceAddedEvent.SliceNr,
		}
		if sliceAddedEvent.SliceNr == 0 {
			handleOutput(ffmpeg, sliceCompletedEvent)
		}
		go listenForCancelMessage(ffmpeg, sliceAddedEvent)
		err := waitForProcessToFinish(ffmpeg, sliceAddedEvent)
		api.ResetProgressMetrics()
		if err == nil {
			return
		} else {
			log.Warn(err.Error())
		}
	}
}

func handleOutput(ffmpeg *cmd.Cmd, slice *entities.SliceCompletedEvent) {
	go process.CaptureOutputLines(ffmpeg,
		func(stdOutLine *string) {
			slice.StdStreams = append(slice.StdStreams, entities.StdStream{
				Line: *stdOutLine,
				FD:   entities.StdOutFileDescriptor,
			})
		}, func(stdErrLine *string) {
			slice.StdStreams = append(slice.StdStreams, entities.StdStream{
				Line: *stdErrLine,
				FD:   entities.StdErrFileDescriptor,
			})
		})
}

func waitForProcessToFinish(ffmpeg *cmd.Cmd, slice *entities.SliceAddedEvent) error {
	status := <-ffmpeg.Start()
	if slice.SliceNr == 0 {
		process.WaitForOutput(ffmpeg)
	}
	if status.Error != nil || status.Exit > 0 {
		if status.Exit < 255 {
			slice.SetComplete(entities.IncompleteAndRequeue)
			return errors.New(fmt.Sprintf("Task failed with exit code %d.", status.Exit))
		} else {
			slice.SetComplete(entities.Complete)
			log.Infof("Task cancelled.")
			return nil
		}
	} else {
		slice.SetComplete(entities.Complete)
		sendSliceCompletedMessage(slice)
		log.Infof("Task slice finished.")
		return nil
	}
}

func handleTaskCancelledEvent(event *entities.TaskCancelledEvent) {
	log.Info(event)
	taskCancelledChan <- event
}

func listenForCancelMessage(ffmpeg *cmd.Cmd, currentTask *entities.SliceAddedEvent) {
	for {
		event := <-taskCancelledChan
		if event.JobID == currentTask.JobID {
			log.Warnf("Cancelling task: %s", event.JobID)
			ffmpeg.Stop()
		} else {
			log.Debugf("TaskCancelledEvent %s does not match current job %s", event.JobID, currentTask.JobID)
		}
		event.SetComplete(entities.Complete)
	}
}

func sendSliceCompletedMessage(slice *entities.SliceAddedEvent) {
	xml, err := entities.ToXml(entities.SliceCompletedEvent{
		SliceNr: slice.SliceNr,
		JobID:   slice.JobID,
	})
	if err == nil {
		service.Publish(sliceCompleteChannel, xml)
	} else {
		log.Warn(err)
	}
}
