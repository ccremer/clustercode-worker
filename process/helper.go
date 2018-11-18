package process

import (
	_ "github.com/aellwein/slf4go-native-adaptor"
	"github.com/ccremer/clustercode-worker/util"
	"github.com/go-cmd/cmd"
	"github.com/micro/go-config"
	"os/exec"
	"time"
)

func StartProcess(args []string) *cmd.Cmd {
	path, err := exec.LookPath("ffmpeg")
	util.PanicOnError(err)
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

func WaitForOutput(ffmpeg *cmd.Cmd) {
	for len(ffmpeg.Stdout) > 0 || len(ffmpeg.Stderr) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
}
