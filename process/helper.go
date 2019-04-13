package process

import (
	"github.com/ccremer/clustercode-worker/config"
	"github.com/go-cmd/cmd"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"time"
)

type (
	Process struct {
		Cmd    *cmd.Cmd
		Args   []string
		config *config.ConfigMap
	}
)

func New(args []string, cfg config.ConfigMap) *Process {
	p := Process{
		Args:   args,
		config: &cfg,
	}
	p.StartProcess()
	return &p
}

func (p *Process) StartProcess() {
	path, err := exec.LookPath("ffmpeg")
	log.Fatal(err)
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}
	extraArgs := p.config.Api.Ffmpeg.DefaultArgs

	ffmpeg := cmd.NewCmdOptions(cmdOptions, path, p.replaceFields(append(p.Args, extraArgs[:]...))[:]...)
	p.Cmd = ffmpeg
	log.WithField("args", ffmpeg.Args).Info("Starting process.", ffmpeg.Args)
	ffmpeg.Start()
}

func (p *Process) PrintOutputLines() {
	for {
		select {
		case line := <-p.Cmd.Stdout:
			log.Debug(line)
			//println(line)
		case line := <-p.Cmd.Stderr:
			log.Debug(line)
			//println(line)
		}
	}
}

func (p *Process) CaptureOutputLines(stdOutCallback func(*string), stdErrCallback func(*string)) {
	for {
		select {
		case line := <-p.Cmd.Stdout:
			stdOutCallback(&line)
		case line := <-p.Cmd.Stderr:
			stdErrCallback(&line)
		}
	}
}

func (p *Process) WaitForOutput() {
	for len(p.Cmd.Stdout) > 0 || len(p.Cmd.Stderr) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
}
