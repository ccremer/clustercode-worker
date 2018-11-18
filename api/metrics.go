package api

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	frameCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "clustercode_compute_frame_count",
		Help: "Current frame counter of a the slice.",
	})
	bitrateDimension = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "clustercode_compute_bitrate",
		Help: "Current bit rate of a the slice.",
	})
	speedDimension = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "clustercode_compute_speed",
		Help: "Current speed of a the slice as factor of the playback time of the media.",
	})
	fpsDimension = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "clustercode_compute_fps",
		Help: "Current frames per second of a the slice.",
	})
	MetricsChan = make(chan Progress)
)

func init() {
	// Register the summary and the histogram with Prometheus's default registry.
	prometheus.MustRegister(frameCounter)
	prometheus.MustRegister(bitrateDimension)
	prometheus.MustRegister(speedDimension)
	prometheus.MustRegister(fpsDimension)
	go func() {
		for {
			select {
			case p := <-MetricsChan:
				SetProgressMetric(p)
			}
		}
	}()
}

func SetProgressMetric(p Progress) {
	frameCounter.Set(float64(p.Frame))
	bitrateDimension.Set(p.Bitrate)
	speedDimension.Set(p.Speed)
	fpsDimension.Set(p.FPS)
}

func ResetProgressMetrics() {
	MetricsChan <- Progress{}
}
