package config

import (
	"bytes"
	"fmt"
	"github.com/ccremer/clustercode-worker/messaging"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	SetupFlags()
	flag.Parse()
}

func LoadConfig() error {
	//viper.Debug()

	defaults := CreateDefaultConfig()
	yml, err := yaml.Marshal(defaults)
	if err != nil {
		log.Warn("Cannot load default config")
	} else {
		viper.SetConfigType("yml")
		_ = viper.ReadConfig(bytes.NewBuffer(yml))
	}

	viper.SetEnvPrefix("CC")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	fullFilename := os.Getenv("CC_CONFIG")
	fullFilenameFromFlag := flag.Lookup("config").Value.String()

	if fullFilenameFromFlag != "" {
		fullFilename = fullFilenameFromFlag
	}

	if fullFilename != "" {

		log.WithField("config", fullFilename).Info("Loading configuration...")

		file := filepath.Base(fullFilename)
		viper.SetConfigName(file)
		viper.AddConfigPath(filepath.Dir(fullFilename))
		viper.AddConfigPath(".")
		viper.AddConfigPath(filepath.Join("/", "usr", "share", "clustercode"))

		return viper.ReadInConfig()
	}

	return nil
}

func GetConfig() ConfigMap {
	cfg := CreateDefaultConfig()
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

func SaveConfigIfRequested() {
	cfg := GetConfig()

	d, err := yaml.Marshal(cfg)
	if err != nil {
		log.WithError(err).Fatal("Could not save config.")
	}
	fileName := flag.Lookup("save-config").Value.String()
	if fileName == "" {
		return
	}

	if err := ioutil.WriteFile(fileName, d, 0665); err != nil {
		log.WithError(err).Fatal("Could not save config.")
	}
	log.WithField("file", fileName).Info("Config saved.")
	os.Exit(0)
}

func createDefaultQosConfig() *messaging.QosOptions {
	return &messaging.QosOptions{
		Enabled:       true,
		PrefetchCount: 1,
	}
}

func createDefaultExchangeConfig(exchangeName string, durable bool) *messaging.ExchangeOptions {
	c := messaging.NewExchangeOptions()
	c.Durable = durable
	c.ExchangeName = exchangeName
	return c
}

func ConfigureLogging() {
	cfg := GetConfig()

	switch cfg.Log.Formatter {
	case LogFormatterJson:
		log.SetFormatter(&log.JSONFormatter{
			DisableTimestamp: cfg.Log.Timestamps,
		})
	case LogFormatterText:
		log.SetFormatter(&log.TextFormatter{
			DisableTimestamp:       cfg.Log.Timestamps,
			FullTimestamp:          true,
			DisableLevelTruncation: true,
		})
		log.WithField("help", "If using ELK/EFK stack, you may want to use the json format (configurable).").
			Info("Using text formatter for logging.")
	default:
		log.WithFields(log.Fields{
			"variable": "log.formatter",
			"help":     fmt.Sprintf("allowed: %s", []string{LogFormatterJson, LogFormatterText}),
			"value":    cfg.Log.Formatter,
			"default":  CreateDefaultConfig().Log.Formatter,
		}).Warnf("Log formatter is not supported. Using default.")
	}

	log.SetOutput(os.Stdout)
	log.SetReportCaller(cfg.Log.Caller)

	level, err := log.ParseLevel(cfg.Log.Level)
	if err != nil {
		log.WithField("error", err).Warn("Using info level.")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
}

func ConvertToChannelConfig(cfg ChannelMap) *messaging.ChannelConfig {
	return &messaging.ChannelConfig{
		QueueOptions:    &cfg.Queue,
		QosOptions:      &cfg.Qos,
		ExchangeOptions: &cfg.Exchange,
	}
}
