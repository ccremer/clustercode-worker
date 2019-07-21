package api

import (
	"fmt"
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/entities"
	"github.com/ccremer/clustercode-worker/messaging"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"net/http"
	"os"
)

type (
	HealthCheckStatus string
	HealthCheckDto struct {
		Checks  []HealthCheck     `json:"checks"`
		Outcome HealthCheckStatus `json:"outcome"`
	}
	HealthCheck struct {
		Id     string                 `json:"id"`
		Status HealthCheckStatus      `json:"status"`
		Data   map[string]interface{} `json:"data"`
	}
	Instance struct {
		config           config.ConfigMap
		MessagingService *messaging.RabbitMqService
	}
)

const (
	UpKey   HealthCheckStatus = "UP"
	DownKey HealthCheckStatus = "DOWN"
)

func (i *Instance) handleLiveness(w http.ResponseWriter, r *http.Request) {
	dto := HealthCheckDto{
		Outcome: UpKey,
		Checks:  []HealthCheck{},
	}

	dto.Checks = append(dto.Checks, checkOutputDir(i.config.Output.Dir))
	dto.Checks = append(dto.Checks, checkInputDir(i.config.Input.Dir))
	dto.Checks = append(dto.Checks, checkMessagingService(i.MessagingService))

	respondHealthRequest(w, r, &dto)
}

func (i *Instance) handleReadiness(w http.ResponseWriter, r *http.Request) {
	dto := HealthCheckDto{
		Outcome: UpKey,
		Checks:  []HealthCheck{},
	}

	dto.Checks = append(dto.Checks, checkMessagingService(i.MessagingService))

	respondHealthRequest(w, r, &dto)
}

func respondHealthRequest(w http.ResponseWriter, r *http.Request, dto *HealthCheckDto) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	failed := funk.Contains(
		funk.Map(dto.Checks, func(check HealthCheck) bool {
			return check.Status == DownKey
		}),
		true)

	if failed {
		dto.Outcome = DownKey
		w.WriteHeader(500)
	} else {
		w.WriteHeader(200)
	}

	json, _ := entities.ToJson(dto)
	_, err := fmt.Fprint(w, json)
	logEvent := log.WithFields(log.Fields{
		"response": json,
		"request":  r.RequestURI,
	})
	if err != nil {
		logEvent.WithError(err).Warn("Could not write response to client.")
	}

	logEvent.Debug("Served request.")
}

func checkOutputDir(dir string) HealthCheck {
	name := dir + "/.health"
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0666)
	if file != nil {
		file.Close()
		os.Remove(name)
	}
	check := HealthCheck{
		Id: "output_dir",
		Data: map[string]interface{}{
			"file": name,
		},
		Status: UpKey,
	}
	if err == nil {
		check.Data["writable"] = true
	} else {
		check.Data["writable"] = false
		check.Status = DownKey
		check.Data["error"] = err.Error()
		log.Warnf("%s", err)
	}
	return check
}

func checkInputDir(path string) HealthCheck {
	check := HealthCheck{
		Id: "input_dir",
		Data: map[string]interface{}{
			"file": path,
		},
		Status: UpKey,
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		check.Status = DownKey
		check.Data["error"] = err.Error()
	}
	return check
}

func checkMessagingService(service *messaging.RabbitMqService) HealthCheck {
	check := HealthCheck{
		Id: "rabbitmq",
		Data: map[string]interface{}{
			"connected": service.IsConnected(),
		},
		Status: UpKey,
	}
	if !service.IsConnected() {
		check.Status = DownKey
	}
	return check
}
