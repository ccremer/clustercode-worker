package api

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestParseProgressOutput_ShouldReturnStruct(t *testing.T) {

	sample := `
frame=4998
fps=383.4
stream_0_0_q=28.0
bitrate=   7.1kbits/s
total_size=48
out_time_ms=197800078
out_time=00:03:17.800078
up_frames=0
drop_frames=0
speed=15.2x
progress=continue

`

	expected := Progress{
		Frame:   4998,
		FPS:     383.4,
		Bitrate: 7.1,
		Speed:   15.2,
	}

	result := parseProgressOutput(sample)
	assert.Equal(t, expected, result)

}
