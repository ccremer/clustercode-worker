package process

import (
	"github.com/go-cmd/cmd"
	"github.com/micro/go-config"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"time"
)

func StartProcess(args []string) *cmd.Cmd {
	path, err := exec.LookPath("ffmpeg")
	log.Fatal(err)
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}
	extraArgs := config.Get("api", "ffmpeg", "defaultargs").StringSlice([]string{"-y", "-hide_banner"})

	ffmpeg := cmd.NewCmdOptions(cmdOptions, path, replaceFields(append(args, extraArgs[:]...))[:]...)
	log.Infof("Starting process: %s", ffmpeg.Args)
	ffmpeg.Start()
	return ffmpeg
}

func PrintOutputLines(cmd *cmd.Cmd) {
	for {
		select {
		case line := <-cmd.Stdout:
			log.Debug(line)
			//println(line)
		case line := <-cmd.Stderr:
			log.Debug(line)
			//println(line)
		}
	}
}

func CaptureOutputLines(cmd *cmd.Cmd, stdOutCallback func(*string), stdErrCallback func(*string)) {
	for {
		select {
		case line := <-cmd.Stdout:
			stdOutCallback(&line)
		case line := <-cmd.Stderr:
			stdErrCallback(&line)
		}
	}
}

func WaitForOutput(ffmpeg *cmd.Cmd) {
	for len(ffmpeg.Stdout) > 0 || len(ffmpeg.Stderr) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
}
