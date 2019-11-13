package api

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	namespace    = "clustercode"
	frameCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "compute_frame_count",
		Help:      "Current frame counter of a the slice.",
	})
	bitrateDimension = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "compute_bitrate",
		Help:      "Current bit rate of a the slice.",
	})
	speedDimension = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "compute_speed",
		Help:      "Current speed of a the slice as factor of the playback time of the media.",
	})
	fpsDimension = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "compute_fps",
		Help:      "Current frames per second of a the slice.",
	})
	metricsChan = make(chan Progress, 2)
)

func EnableMetrics() {
	// Register the summary and the histogram with Prometheus's default registry.
	prometheus.MustRegister(frameCounter)
	prometheus.MustRegister(bitrateDimension)
	prometheus.MustRegister(speedDimension)
	prometheus.MustRegister(fpsDimension)
	go func() {
		for {
			select {
			case p := <-metricsChan:
				SetProgressMetric(p)
			default:
				continue
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
	metricsChan <- Progress{}
}
