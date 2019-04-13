package api

import (
	"fmt"
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func StartHttpServer(service *messaging.RabbitMqService) {

	cfg := config.GetConfig()
	addr := cfg.Api.Http.Address
	monitoringInstance := Instance{
		config:cfg,
		MessagingService: service,
	}
	log.WithField("address", addr).Info("Starting http server")
	http.HandleFunc("/", handleRoot)
	http.HandleFunc(cfg.Api.Http.HealthUri, monitoringInstance.handleHealth)
	http.HandleFunc(cfg.Api.Http.ReadyUri, monitoringInstance.handleReadyness)
	if cfg.Prometheus.Enabled {
		EnableMetrics()
		http.Handle(cfg.Prometheus.Uri, promhttp.Handler())
	}
	go func() {
		err := http.ListenAndServe(addr, nil)
		log.Fatal(err)
	}()

	ffmpegProtocol := strings.ToLower(cfg.Api.Ffmpeg.Protocol)
	switch ffmpegProtocol {
	case "unix":
		openUnixSocket(cfg.Api.Ffmpeg.Unix)
	default:
		log.WithField("protocol", ffmpegProtocol).Fatal("ffmpeg protocol is not supported")
	}

}

func handleRoot(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "This page is intentionally left blank. You might want to check the API URLs")
}
