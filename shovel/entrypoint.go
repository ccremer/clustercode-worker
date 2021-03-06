package shovel

import "github.com/ccremer/clustercode-worker/messaging"

func NewShovelInstance(service messaging.Service) Instance {
	i := Instance{
		MessagingService:   service,
		taskCompletedQueue: getTaskCompletedQueue(),
	}
	service.AddChannelConfig(i.taskCompletedQueue)
	service.AddChannelConfig(createTaskAddedQueue(i.ProcessTask))
	return i
}
