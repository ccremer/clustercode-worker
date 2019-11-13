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
		CurrentTask          *entities.SliceAddedEvent
		CurrentProcess       *process.Process
	}
	TaskResult struct {
		SliceAddedEvent *entities.SliceAddedEvent
		ErrorCode       int
		Error           error
		Cancelled       bool
	}
)

func (i *Instance) ProcessSlice(sliceAddedEvent *entities.SliceAddedEvent) TaskResult {
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
	i.CurrentProcess = ffmpeg
	i.CurrentTask = sliceAddedEvent
	sliceCompletedEvent := &entities.SliceCompletedEvent{
		JobID:   sliceAddedEvent.JobID,
		SliceNr: sliceAddedEvent.SliceNr,
	}
	if sliceAddedEvent.SliceNr == 0 {
		i.handleOutput(ffmpeg, sliceCompletedEvent)
	}
	result := TaskResult{SliceAddedEvent: sliceAddedEvent}
	waitForProcessToFinish(ffmpeg, sliceAddedEvent, &result)
	api.ResetProgressMetrics()
	i.CurrentProcess = nil
	i.CurrentProcess = nil
	return result
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

func waitForProcessToFinish(ffmpeg *process.Process, slice *entities.SliceAddedEvent, result *TaskResult) {
	status := <-ffmpeg.Cmd.Start()
	if slice.SliceNr == 0 {
		ffmpeg.WaitForOutput()
	}
	if status.Error != nil || status.Exit > 0 {
		if status.Exit < 255 {
			result.Error = errors.New(fmt.Sprintf("Task failed with exit code %d.", status.Exit))
		} else {
			log.Infof("Task cancelled.")
			result.Cancelled = true
		}
	} else {
		log.Infof("Task slice finished.")
	}
}

func (i *Instance) CancelTask(event *entities.TaskCancelledEvent) bool {
	logEntry := log.WithField("task_id", event.JobID)
	if event.JobID == i.CurrentTask.JobID {
		logEntry.Debug("Cancelling task.")
		err := i.CurrentProcess.Cmd.Stop()
		if err != nil {
			logEntry.
				WithError(err).
				Warn("Task cancelled before it has started.")
		}
		return true
	} else {
		logEntry.
			WithField("current", i.CurrentTask.JobID).
			Debug("TaskCancelledEvent does not match current job.")
		return false
	}
}

func (i *Instance) sendSliceCompletedMessage(slice *entities.SliceAddedEvent) {
	payload, err := entities.ToJson(entities.SliceCompletedEvent{
		SliceNr: slice.SliceNr,
		JobID:   slice.JobID,
	})
	if err == nil {
		i.MessagingService.Publish(i.SliceCompleteChannel, payload)
	} else {
		log.Warn(err)
	}
}
