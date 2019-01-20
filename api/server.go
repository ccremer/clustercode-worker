package api

import (
	"fmt"
	"github.com/micro/go-config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func StartServer() {

	port := config.Get("http", "port").String("8080")
	log.Infof("Starting http server on port %s", port)
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/health", handleHealth)
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":"+port, nil)
		log.Fatal(err)
	}()

	ffmpegProtocol := strings.ToLower(config.Get("api", "ffmpeg", "protocol").String("unix"))
	switch ffmpegProtocol {
	case "unix":
		openUnixSocket()
	default:
		log.WithField("protocol", ffmpegProtocol).Fatal("protocol is not supported")
	}

}

func handleRoot(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "This page is intentionally left blank. You might want to check /health")
}
