package api

import (
	"fmt"
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestSetProgressMetric_ShouldUpdateMetrics(t *testing.T) {

	if testing.Short() {
		t.Skipf("Skipping integration test")
	}

	if err := config.LoadConfig(); err != nil {
		t.Fatalf("Could not load config: %s", err)
	}

	StartHttpServer(&messaging.RabbitMqService{})
	time.Sleep(1 * time.Second)
	expected := Progress{
		FPS:     20.4,
		Bitrate: 234.3,
		Speed:   10.6,
		Frame:   1230,
	}

	SetProgressMetric(expected)

	addr := config.GetConfig().Api.Http.Address
	uri := config.GetConfig().Prometheus.Uri

	resp, err := http.Get(fmt.Sprintf("http://localhost%s/%s", addr, uri))
	assert.Empty(t, err)
	defer resp.Body.Close()
	linesRaw, err := ioutil.ReadAll(resp.Body)
	assert.Empty(t, err)
	lines := string(linesRaw)
	assert.Contains(t, lines, "clustercode_compute_bitrate 234.3")
	assert.Contains(t, lines, "clustercode_compute_fps 20.4")
	assert.Contains(t, lines, "clustercode_compute_frame_count 1230")
	assert.Contains(t, lines, "clustercode_compute_speed 10.6")

}
