package process

import (
	"github.com/ccremer/clustercode-worker/config"
	"github.com/go-cmd/cmd"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"time"
)

type (
	Process struct {
		Cmd           *cmd.Cmd
		Args          []string
		DefaultArgs   []string
		StdOutHandler func(*string)
		StdErrHandler func(*string)
	}
)

var (
	PrintToStdOutHandler = func(line *string) {
		_, _ = os.Stdout.WriteString(*line + "\n")
	}
	PrintToStdErrHandler = func(line *string) {
		_, _ = os.Stderr.WriteString(*line + "\n")
	}
	PrintToDebugHandler = func(line *string) {
		log.Debug(*line)
	}
	NullHandler = func(line *string) {}
)

func New(args []string, cfg config.ConfigMap) *Process {
	p := Process{
		Args:        args,
		DefaultArgs: cfg.Api.Ffmpeg.DefaultArgs,
	}
	return &p
}

func (p *Process) StartProcess(replacements map[string]string) <-chan cmd.Status {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Fatal(err)
	}
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}

	ffmpeg := cmd.NewCmdOptions(cmdOptions, path, p.replaceFields(
		append(p.Args, p.DefaultArgs[:]...),
		replacements)[:]...)
	p.Cmd = ffmpeg
	log.WithField("args", strings.Join(ffmpeg.Args, "','")).Info("Starting process.")
	status := ffmpeg.Start()
	go p.CaptureOutputLines(p.StdOutHandler, p.StdErrHandler)
	p.WaitForOutput()
	return status
}

func (p *Process) SetBothOutHandlers(handler func(*string)) {
	p.StdErrHandler = handler
	p.StdOutHandler = handler
}

func (p *Process) CaptureOutputLines(stdOutCallback func(*string), stdErrCallback func(*string)) {
	var outHandler = NullHandler
	var errHandler = NullHandler
	if stdOutCallback != nil {
		outHandler = stdOutCallback
	}
	if stdErrCallback != nil {
		errHandler = stdErrCallback
	}
	for {
		select {
		case line := <-p.Cmd.Stdout:
			outHandler(&line)
		case line := <-p.Cmd.Stderr:
			errHandler(&line)
		}
	}
}

func (p *Process) WaitForOutput() {
	for len(p.Cmd.Stdout) > 0 || len(p.Cmd.Stderr) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
}
