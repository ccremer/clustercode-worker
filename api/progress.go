package api

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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

var (
	// original: \s*(\d+\.?\d*).*
	re        = regexp.MustCompile("\\s*(\\d+\\.?\\d*).*")
	extraArgs []string
)

func openUnixSocket(path string) {

	os.Remove(path)
	l, err := net.Listen("unix", path)
	if err != nil {
		log.Fatal(err)
	}
	extraArgs = []string{"-progress", fmt.Sprintf("unix://%s", path)}
	go func() {
		for {
			fd, err := l.Accept()
			if err != nil {
				log.Panic(err)
			}
			go readSocket(fd)
		}
	}()
}

func GetExtraArgsForProgress() []string {
	return extraArgs
}

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

	if fields["progress"] == "end" {
		return Progress{}
	} else {
		result := Progress{
			Frame:   frame,
			FPS:     math.Round(fps*100) / 100,
			Bitrate: math.Round(bitrate*100) / 100,
			Speed:   math.Round(speed*100) / 100,
		}
		return result
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
		metricsChan <- parseProgressOutput(string(data))
	}
}
