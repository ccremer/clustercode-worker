package api

import (
	"github.com/ccremer/clustercode-worker/util"
	"github.com/micro/go-config"
	"math"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type (
	Progress struct {
		Frame   int
		FPS     float64
		Bitrate float64
		Speed   float64
	}
)

func openUnixSocket() {

	path := config.Get("api", "ffmpeg", "unix").String("/tmp/ffmpeg.sock")

	os.Remove(path)
	l, err := net.Listen("unix", path)
	if err != nil {
		util.PanicOnError(err)
	}

	for {
		fd, err := l.Accept()
		if err != nil {
			util.PanicOnError(err)
		}
		go readSocket(fd)
	}
}

// original: \s*(\d+\.?\d*).*
var re = regexp.MustCompile("\\s*(\\d+\\.?\\d*).*")

func parseProgressOutput(out string) Progress {
	fields := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		fragments := strings.Split(line, "=")
		if len(fragments) > 1 {
			fields[fragments[0]] = fragments[1]
		}
	}
	frame, _ := strconv.Atoi(fields["frame"])
	fps, _ := strconv.ParseFloat(fields["fps"], 32)

	match := re.FindStringSubmatch(fields["bitrate"])
	bitrate, _ := strconv.ParseFloat(match[1], 32)

	speedRaw := fields["speed"]
	speed, _ := strconv.ParseFloat(speedRaw[0:len(speedRaw)-1], 32)

	return Progress{
		Frame:   frame,
		FPS:     math.Round(fps*100) / 100,
		Bitrate: math.Round(bitrate*100) / 100,
		Speed:   math.Round(speed*100) / 100,
	}
}

func readSocket(c net.Conn) {
	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}

		data := buf[0:nr]
		log.Infof("%v", parseProgressOutput(string(data)))
	}
}
