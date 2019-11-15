package messaging

type Service interface {
	Start(configs ...*ChannelConfig)
	IsConnected() bool
	Publish(config *ChannelConfig, payload string)
	AddChannelConfig(config *ChannelConfig)
}
