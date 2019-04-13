package api

import (
	"fmt"
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/entities"
	"github.com/ccremer/clustercode-worker/messaging"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type (
	HealthCheckDto struct {
		InputDir  string `json:"input_dir"`
		OutputDir string `json:"output_dir"`
		Messaging string `json:"messaging"`
	}
	ReadynessCheckDto struct {
		Database string `json:"database"`

	}
	Instance struct {
		config config.ConfigMap
		MessagingService *messaging.RabbitMqService
	}
)

func (i *Instance) handleHealth(w http.ResponseWriter, r *http.Request) {
	dto := HealthCheckDto{}
	faulty := false
	msg, failure := checkOutputDir(i.config.Output.Dir)
	if failure {
		faulty = true
	}
	dto.OutputDir = msg
	msg, failure = checkInputDir(i.config.Input.Dir)
	if failure {
		faulty = true
	}
	dto.InputDir = msg
	if i.MessagingService.IsConnected() {
		dto.Messaging = "ok"
	} else {
		dto.Messaging = "disconnected"
		faulty = true
	}

	if faulty {
		w.WriteHeader(500)
	} else {
		w.WriteHeader(200)
	}

	json, _ := entities.ToJson(dto)
	fmt.Fprint(w, json)

	log.Debugf("response: %s", json)
}

func (i *Instance) handleReadyness(w http.ResponseWriter, r *http.Request) {

}

func checkOutputDir(dir string) (string, bool) {
	name := dir + "/.health"
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		log.Warnf("%s", err)
		return fmt.Sprint(err), true
	} else {
		os.Remove(name)
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
