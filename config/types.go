package config

import (
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/prometheus/common/log"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func CreateDefaultConfig() ConfigMap {
	return ConfigMap{
		Role: "",
		Log: LogMap{
			Level:      "info",
			Timestamps: false,
			Formatter:  "text",
			Caller:     false,
		},
		RabbitMq: RabbitMqMap{
			Url: "amqp://guest:guest@rabbitmq:5672/",
			Channels: ChannelsCollection{
				Task: TaskMap{
					Added: ChannelMap{
						Queue: messaging.QueueOptions{
							Enabled:   true,
							QueueName: "task-added",
							Durable:   true,
						},
						Qos: *createDefaultQosConfig(),
					},
					Completed: ChannelMap{
						Queue: messaging.QueueOptions{
							Enabled:   true,
							QueueName: "task-completed",
							Durable:   true,
						},
					},
					Cancelled: ChannelMap{
						Queue: messaging.QueueOptions{
							Enabled:   true,
							Exclusive: true,
						},
						Exchange: *createDefaultExchangeConfig("task-cancelled", true),
					},
				},
				Slice: SliceMap{
					Added: ChannelMap{
						Queue: messaging.QueueOptions{
							Enabled:   true,
							QueueName: "slice-added",
							Durable:   true,
						},
						Qos: *createDefaultQosConfig(),
					},
					Completed: ChannelMap{
						Queue: messaging.QueueOptions{
							Enabled:   true,
							QueueName: "slice-completed",
							Durable:   true,
						},
					},
				},
			},
		},
		Api: ApiMap{
			Ffmpeg: FfmpegMap{
				DefaultArgs: []string{"-y", "-hide_banner", "-nostats"},
				Protocol:    "unix",
				Unix:        "/tmp/ffmpeg.sock",
			},
			Http: HttpMap{
				Address:   ":8080",
				ReadyUri:  "/.well-known/ready",
				HealthUri: "/.well-known/health",
			},
			Schema: SchemaMap{
				Path: "/usr/share/clustercode/clustercode_v1.xsd",
			},
		},
		Input: InputMap{
			Dir: "/input",
		},
		Output: OutputMap{
			Dir:    "/output",
			TmpDir: "/var/tmp/clustercode",
		},
		Prometheus: PrometheusMap{
			Enabled: true,
			Uri:     "/.well-known/metrics",
		},
	}
}

func SetupFlags() {

	cfg := CreateDefaultConfig()

	flag.StringP("log.level", "l", cfg.Log.Level,
		"Logging level. Allowed values are either ['debug','info','warn','error','fatal']")
	flag.StringP("config", "c", "",
		"Config file path and name without extension")
	flag.StringP("role", "R", cfg.Role,
		"The role of this worker. Allowed values are either ['compute','shovel']")
	flag.String("rabbitmq.url", cfg.RabbitMq.Url, "RabbitMq connection string")
	flag.String("api.http.address", cfg.Api.Http.Address, "HTTP API server listen address")
	flag.String("api.schema.path", cfg.Api.Schema.Path, "XML Schema definition file for API messages")
	flag.Bool("prometheus.enabled", cfg.Prometheus.Enabled, "Whether metrics exporter is enabled")

	if err := viper.BindPFlags(flag.CommandLine); err != nil {
		log.Fatal(err)
	}
}

type (
	ConfigMap struct {
		Log        LogMap
		RabbitMq   RabbitMqMap
		Api        ApiMap
		Input      InputMap
		Output     OutputMap
		Prometheus PrometheusMap
		Role       string
	}
	LogMap struct {
		Level      string
		Timestamps bool
		Formatter  string
		Caller     bool
	}
	RabbitMqMap struct {
		Url      string
		Channels ChannelsCollection
	}
	ChannelsCollection struct {
		Task  TaskMap
		Slice SliceMap
	}
	ChannelMap struct {
		Queue    messaging.QueueOptions
		Exchange messaging.ExchangeOptions
		Qos      messaging.QosOptions
	}
	TaskMap struct {
		Added     ChannelMap
		Completed ChannelMap
		Cancelled ChannelMap
	}
	SliceMap struct {
		Added     ChannelMap
		Completed ChannelMap
	}
	ApiMap struct {
		Ffmpeg FfmpegMap
		Http   HttpMap
		Schema SchemaMap
	}
	HttpMap struct {
		Address   string
		HealthUri string
		ReadyUri  string
	}
	FfmpegMap struct {
		DefaultArgs []string
		Protocol    string
		Unix        string
	}
	SchemaMap struct {
		Path string
	}
	InputMap struct {
		Dir string
	}
	OutputMap struct {
		Dir    string
		TmpDir string
	}
	PrometheusMap struct {
		Enabled bool
		Uri     string
	}
)
