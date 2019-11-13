package config

import "github.com/ccremer/clustercode-worker/messaging"

func CreateDefaultConfig() ConfigMap {
	return ConfigMap{
		Role: "",
		Log: LogMap{
			Level:      LogLevelInfo,
			Timestamps: true,
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
					"${OUTPUT}"},
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
