package config

import (
	"fmt"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/prometheus/common/log"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"strings"
)

const (
	RoleShovel            = "shovel"
	RoleCompute           = "compute"
	LogLevelDebug         = "debug"
	LogLevelInfo          = "info"
	LogLevelWarn          = "warn"
	LogLevelError         = "error"
	LogLevelFatal         = "fatal"
	LogFormatterText      = "text"
	LogFormatterJson      = "json"
	ApiFfmpegDebugNone    = "none"
	ApiFfmpegDebugLog     = "to-log"
	ApiFfmpegDebugOut     = "to-out"
	ApiFfmpegProtocolUnix = "unix"
)

func CreateDefaultConfig() ConfigMap {
	return ConfigMap{
		Role: "",
		Log: LogMap{
			Level:      LogLevelInfo,
			Timestamps: false,
			Formatter:  LogFormatterText,
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
				Protocol:    ApiFfmpegProtocolUnix,
				Unix:        "/tmp/ffmpeg.sock",
				SplitArgs: []string{
					"-i", "${INPUT}", "-c", "copy", "-map", "0", "-segment_time", "${SLICE_SIZE}", "-f", "segment",
					"${TMP}/${JOB}_segment_%d${FORMAT}"},
				MergeArgs: []string{"-f", "concat", "-i", "concat.txt", "-c", "copy", "movie_out.mkv"},
				Debug:     ApiFfmpegDebugNone,
			},
			Http: HttpMap{
				Address:      ":8080",
				ReadinessUri: "/health/ready",
				LivenessUri:  "/health/live",
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
			Uri:     "/metrics",
		},
	}
}

func SetupFlags() {

	cfg := CreateDefaultConfig()

	flag.StringP("log.level", "l", cfg.Log.Level,
		fmt.Sprintf("Logging level. Allowed values are either [%s]", strings.Join([]string{
			LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelFatal}[:], ", ")))
	flag.StringP("config", "c", "",
		"Config file path and name without extension")
	flag.StringP("role", "R", cfg.Role,
		fmt.Sprintf("The role of this worker. Allowed values are either [%s]", strings.Join([]string{
			RoleCompute, RoleShovel}[:], ", ")))
	flag.String("rabbitmq.url", cfg.RabbitMq.Url, "RabbitMq connection string")
	flag.String("api.http.address", cfg.Api.Http.Address, "HTTP API server listen address")
	flag.Bool("prometheus.enabled", cfg.Prometheus.Enabled, "Whether metrics exporter is enabled")
	flag.String("save-config", "",  "Save the final config to the given file path and exit")

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
	}
	HttpMap struct {
		Address      string
		LivenessUri  string
		ReadinessUri string
	}
	FfmpegMap struct {
		DefaultArgs []string
		Protocol    string
		Unix        string
		SplitArgs   []string
		MergeArgs   []string
		Debug       string
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
