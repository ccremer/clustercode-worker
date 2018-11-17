package api

import (
	"fmt"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/micro/go-config"
	"net/http"
	"os"
)

type HealthCheckDto struct {
	InputDir  string `json:"input_dir"`
	OutputDir string `json:"output_dir"`
	Messaging string `json:"messaging"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	dto := HealthCheckDto{}
	faulty := false
	msg, failure := checkOutputDir(config.Get("output", "dir").String("/output") + "/.clustercode-health")
	if failure {
		faulty = true
	}
	dto.OutputDir = msg
	msg, failure = checkInputDir(config.Get("input", "dir").String("/input") + "/0")
	if failure {
		faulty = true
	}
	dto.InputDir = msg
	dto.Messaging = "unknown"

	if faulty {
		w.WriteHeader(500)
	} else {
		w.WriteHeader(200)
	}

	json, _ := messaging.ToJson(dto)
	fmt.Fprint(w, json)

	log.Debugf("response: %s", json)
}

func checkOutputDir(path string) (string, bool) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		log.Warnf("%s", err)
		return fmt.Sprint(err), true
	} else {
		os.Remove(path)
		return "ok", false
	}
}

func checkInputDir(path string) (string, bool) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Sprint(err), true
	} else {
		return "ok", false
	}
}
