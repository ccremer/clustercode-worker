package api

import (
	"fmt"
	"github.com/aellwein/slf4go"
	"github.com/ccremer/clustercode-worker/util"
	"github.com/micro/go-config"
	"net/http"
	"strings"
)

var log = slf4go.GetLogger("httpapi")

func Init() {
	log.SetLevel(util.StringToLogLevel(config.Get("log", "level").String("info")))
}

func StartServer() {

	port := config.Get("http", "port").String("8080")
	log.Infof("Starting http server on port %s", port)
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/health", handleHealth)
	go func() {
		err := http.ListenAndServe(":"+port, nil)
		util.PanicOnError(err)
	}()

	ffmpegProtocol := strings.ToLower(config.Get("api", "ffmpeg", "protocol").String("unix"))
	switch ffmpegProtocol {
	case "unix":
		openUnixSocket()
	default:
		util.PanicWithMessage("Protocol %s is not supported!", ffmpegProtocol)
	}

}

func handleRoot(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "This page is intentionally left blank. You might want to check /health")
}
