package main

import (
	"fmt"
	"github.com/ccremer/clustercode-worker/api"
	"github.com/ccremer/clustercode-worker/compute"
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/ccremer/clustercode-worker/shovel"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"strconv"
	"strings"
)

var (
	version = "latest"
	commit  = "snapshot"
	date    = "unknown"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.
			WithError(err).
			WithField("help", "Be sure to NOT specify the file extension.").
			Error("Could not load config.")
	}
	if result, _ := strconv.ParseBool(flag.Lookup("version").Value.String()); result {
		fmt.Println(version)
		fmt.Println(commit)
		return
	}
	config.ConfigureLogging()

	cfg := config.GetConfig()
	config.SaveConfigIfRequested()

	role := cfg.Role
	if role != config.RoleShovel && role != config.RoleCompute {
		log.WithFields(log.Fields{
			"variable": "role",
			"help": fmt.Sprintf("Specify the role either in CC_ROLE, by providing the CLI flag --role, "+
				"or in a config file. Allowed values: [%s]",
				strings.Join([]string{config.RoleCompute, config.RoleShovel}, ",")),
		}).Fatal("Role specification of this worker is required.")
	}

	service := messaging.NewRabbitMqService(cfg.RabbitMq.Url)
	api.StartHttpServer(service)

	go func() {
		service.Start()

		if role == config.RoleCompute {
			compute.NewComputeInstance(service)
		} else if role == config.RoleShovel {
			shovel.NewShovelInstance(service)
		}
	}()

	log.WithFields(log.Fields{
		"version":    version,
		"commit":     commit,
		"build_date": date,
		"role":       role,
	}).Info("Startup complete.")
	// Keep it running:
	<-make(chan bool)
}
