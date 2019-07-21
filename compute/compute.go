package compute

import (
	"errors"
	"fmt"
	"github.com/ccremer/clustercode-worker/api"
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/entities"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/ccremer/clustercode-worker/process"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type (
	Instance struct {
		MessagingService     *messaging.RabbitMqService
		SliceCompleteChannel *messaging.ChannelConfig
		sliceCompleteChan    chan *entities.SliceCompletedEvent
		taskCancelledChan    chan *entities.TaskCancelledEvent
	}
)

func NewComputeInstance(service *messaging.RabbitMqService) Instance {
	i := Instance{
		MessagingService:  service,
		sliceCompleteChan: make(chan *entities.SliceCompletedEvent),
		taskCancelledChan: make(chan *entities.TaskCancelledEvent),
	}
	service.AddChannelConfig(getSliceCompleteQueue())
	service.AddChannelConfig(getSliceAddedQueue(i.handleSliceAddedEvent))
	service.AddChannelConfig(getTaskCancelledQueue(i.handleTaskCancelledEvent))

	return i
}

func (i *Instance) handleSliceAddedEvent(sliceAddedEvent *entities.SliceAddedEvent) {
	for {
		log.WithFields(log.Fields{
			"job_id":   sliceAddedEvent.JobID,
			"slice_nr": sliceAddedEvent.SliceNr,
		}).Info("Processing SliceAddedEvent")
		c := config.GetConfig()
		ffmpeg := process.New(append(sliceAddedEvent.Args, api.GetExtraArgsForProgress()[:]...), c)
		ffmpeg.StartProcess(map[string]string{
			"${INPUT}":  c.Input.Dir,
			"${OUTPUT}": c.Output.Dir,
			"${TMP}":    c.Output.TmpDir,
			"${SLICES}": strconv.Itoa(sliceAddedEvent.SliceNr),
		})
		sliceCompletedEvent := &entities.SliceCompletedEvent{
			JobID:   sliceAddedEvent.JobID,
			SliceNr: sliceAddedEvent.SliceNr,
		}
		if sliceAddedEvent.SliceNr == 0 {
			i.handleOutput(ffmpeg, sliceCompletedEvent)
		}
		go i.listenForCancelMessage(ffmpeg, sliceAddedEvent)
		err := i.waitForProcessToFinish(ffmpeg, sliceAddedEvent)
		api.ResetProgressMetrics()
		if err == nil {
			return
		} else {
			log.Warn(err.Error())
		}
	}
}

func (i *Instance) handleOutput(ffmpeg *process.Process, slice *entities.SliceCompletedEvent) {
	go ffmpeg.CaptureOutputLines(
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

func (i *Instance) waitForProcessToFinish(ffmpeg *process.Process, slice *entities.SliceAddedEvent) error {
	status := <-ffmpeg.Cmd.Start()
	if slice.SliceNr == 0 {
		ffmpeg.WaitForOutput()
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
		i.sendSliceCompletedMessage(slice)
		log.Infof("Task slice finished.")
		return nil
	}
}

func (i *Instance) handleTaskCancelledEvent(event *entities.TaskCancelledEvent) {
	log.Info(event)
	i.taskCancelledChan <- event
}

func (i *Instance) listenForCancelMessage(ffmpeg *process.Process, currentTask *entities.SliceAddedEvent) {
	for {
		event := <-i.taskCancelledChan
		logEntry := log.WithField("task_id", event.JobID)
		if event.JobID == currentTask.JobID {
			logEntry.Warn("Cancelling task.")
			ffmpeg.Cmd.Stop()
		} else {
			logEntry.
				WithField("current", currentTask.JobID).
				Debug("TaskCancelledEvent does not match current job.")
		}
		event.SetComplete(entities.Complete)
	}
}

func (i *Instance) sendSliceCompletedMessage(slice *entities.SliceAddedEvent) {
	xml, err := entities.ToJson(entities.SliceCompletedEvent{
		SliceNr: slice.SliceNr,
		JobID:   slice.JobID,
	})
	if err == nil {
		i.MessagingService.Publish(i.SliceCompleteChannel, xml)
	} else {
		log.Warn(err)
	}
}
