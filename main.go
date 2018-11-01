package main

import (
    "github.com/aellwein/slf4go"
    _ "github.com/aellwein/slf4go-native-adaptor"
    "github.com/ccremer/clustercode-worker/health"
    "github.com/ccremer/clustercode-worker/util"
    "github.com/go-cmd/cmd"
    "os"
    "os/exec"
    "strings"
    "time"
)

var log slf4go.Logger

func main() {

    log = slf4go.GetLogger("main")

    health.StartServer()

    computeRole := "compute"
    shovelRole := "shovel"

    role := os.Getenv("CC_ROLE")
    if role == computeRole {

    } else if role == shovelRole {

    } else {
        util.PanicWithMessage("You need to specify this worker's role using CC_ROLE to one of %s",
            computeRole, shovelRole)
    }

    path, err := exec.LookPath("ffmpeg")
    util.PanicOnError(err)
    log.Infof("ffmpeg executable is in %s", path)

    command := path + " -hide_banner -y -t 20 -s 640x480 -f rawvideo -pix_fmt rgb24 -r 25 -i /dev/zero vendor/empty.mpeg"
    log.Infof(command)

    cmdArgs := strings.Fields(command)

    cmdOptions := cmd.Options{
        Buffered:  false,
        Streaming: true,
    }

    ffmpeg := cmd.NewCmdOptions(cmdOptions, cmdArgs[0], cmdArgs[1:]...)

    ffmpeg.Start()

    go func() {
        for {
            select {
            case line := <-ffmpeg.Stdout:
                log.Info(line)
            case line := <-ffmpeg.Stderr:
                log.Error(line)
            }
        }
    }()

    // Run and wait for Cmd to return, discard Status
    <-ffmpeg.Start()

    // Cmd has finished but wait for goroutine to print all lines
    for len(ffmpeg.Stdout) > 0 || len(ffmpeg.Stderr) > 0 {
        time.Sleep(10 * time.Millisecond)
    }
}
