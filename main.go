package main

import (
	"fmt"
	"github.com/ccremer/clustercode-api-gateway/entities"
	"github.com/ccremer/clustercode-api-gateway/messaging"
	"github.com/ccremer/clustercode-api-gateway/schema"
	"github.com/ccremer/clustercode-worker/api"
	"github.com/ccremer/clustercode-worker/compute"
	"github.com/ccremer/clustercode-worker/shovel"
	"github.com/micro/go-config"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

func main() {
	fmt.Println("started!")

	SetupFlags()
	entities.Validator = &schema.Validator{}
	entities.Validator.LoadXmlSchema("../clustercode-api-gateway/schema/clustercode_v1.xsd")
	displayHelp := flag.Bool("help", false, "Displays help text and exits")
	flag.Parse()
	if *displayHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	log.Infof("Loading configuration...")
	LoadConfig()

	service := messaging.NewRabbitMqService("amqp://guest:guest@localhost:5672/")
	service.Start()

	computeRole := "compute"
	shovelRole := "shovel"
	role := config.Get("role").String("compute")
	if role == computeRole {
		log.Infof("Enable role: %s", computeRole)
		compute.Start(service)
	} else if role == shovelRole {
		log.Infof("Enable role: %s", shovelRole)
		shovel.Start()
	} else {
		log.WithFields(log.Fields{
			"variable": "CC_ROLE",
			"allowed":  []string{computeRole, shovelRole},
		}).Fatal("role specification of this worker is required")
	}

	api.StartServer()

	forever := make(chan bool)
	log.Infof("Startup complete")
	<-forever
}

func SetupFlags() map[string]interface{} {
	m := make(map[string]interface{})

	m["log.level"] = *flag.StringP("log-level", "l", "info", "Log level")
	return m
}

func LoadConfig() {
	viper.Debug()
	viper.SetConfigName("clustercode")
	viper.AddConfigPath(".")
	//viper.AddConfigPath("")
	//viper.G
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	v := viper.GetViper()
	fmt.Println(v)
	//v.M
}

func ConfigureLogging() {
	key := "log"
	disableTimestamps := !config.Get(key, "timestamps").Bool(false)
	formatter := config.Get(key, "formatter").String("json")
	switch formatter {
	case "json":
		log.SetFormatter(&log.JSONFormatter{DisableTimestamp: disableTimestamps})
	case "text":
		log.SetFormatter(&log.TextFormatter{DisableTimestamp: disableTimestamps, FullTimestamp: true})
	default:
		log.Warnf("Log formatter '%s' is not supported. Using default", formatter)
	}

	log.SetOutput(os.Stdout)
	log.SetReportCaller(config.Get(key, "caller").Bool(false))

	level, err := log.ParseLevel(config.Get(key, "level").String("info"))
	if err != nil {
		log.Warnf("%s. Using info.", err)
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
}
