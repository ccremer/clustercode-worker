package shovel

import (
	"errors"
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/entities"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/ccremer/clustercode-worker/process"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strconv"
)

type (
	Instance struct {
		MessagingService   messaging.Service
		config             config.ConfigMap
		taskCompletedQueue *messaging.ChannelConfig
	}
	TaskResult struct {
		TaskAddedEvent   *entities.TaskAddedEvent
		Error            error
		ExitCode         int
		SliceCountResult int
		TaskType         entities.TaskType
	}
)

func (i *Instance) ProcessTask(task *entities.TaskAddedEvent) TaskResult {
	logEntry := log.WithField("task_id", task.JobID)
	logEntry.Info("Processing task.")
	if task.Type == "SPLIT" {
		return i.splitMedia(task)
	} else if task.Type == "MERGE" {
		return i.mergeMedia(task)
	}
	return TaskResult{Error: errors.New("task type is undefined")}
}

func (i *Instance) attachStdListeners(c config.ConfigMap, p *process.Process) {
	switch c.Api.Ffmpeg.Debug {
	case config.ApiFfmpegDebugLog:
		if log.IsLevelEnabled(log.DebugLevel) {
			p.SetBothOutHandlers(process.PrintToDebugHandler)
		}
	case config.ApiFfmpegDebugOut:
		p.StdOutHandler = process.PrintToStdOutHandler
		p.StdErrHandler = process.PrintToStdErrHandler
	}
}

func (i *Instance) splitMedia(task *entities.TaskAddedEvent) TaskResult {
	logEntry := log.WithField("task_id", task.JobID)
	c := config.GetConfig()

	args := c.Api.Ffmpeg.SplitArgs
	p := process.New(args, c)
	i.attachStdListeners(c, p)

	status := <-p.StartProcess(map[string]string{
		"${INPUT}": task.Media.GetSubstitutedPath(c.Input.Dir),
		"${OUTPUT}": filepath.Join(
			c.Output.TmpDir,
			task.JobID+"_segment_%d"+filepath.Ext(task.Media.RelativePath())),
		"${SLICE_SIZE}": strconv.Itoa(task.SliceSize),
	})

	result := TaskResult{
		TaskAddedEvent: task,
		TaskType:       entities.Split,
	}

	logEntry = logEntry.WithField("exit_code", status.Exit)
	if status.Error != nil || status.Exit > 0 {
		logEntry.Info("Task failed.")
		result.Error = status.Error
		result.ExitCode = status.Exit
		return result
	}

	matches, err := filepath.Glob(filepath.Join(c.Output.TmpDir, task.JobID+"_segment_*"))
	if err == nil {
		sliceCount := len(matches)
		log.WithField("sliceCount", sliceCount).Info("Task finished.")
		result.SliceCountResult = sliceCount
	} else {
		log.WithError(err).Panic("Bad pattern.")
	}
	return result
}

func (i *Instance) mergeMedia(event *entities.TaskAddedEvent) TaskResult {
	// TODO: finish
	return TaskResult{}
}
