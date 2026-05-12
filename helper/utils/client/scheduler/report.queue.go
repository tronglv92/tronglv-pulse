package scheduler

import (
	"context"
	"pulse/helper/utils/queue"
)

type queueReportClient struct {
	topicName      string
	producerClient queue.Producer
}

func NewQueueReportClient(producerClient queue.Producer, topicName string) JobReporter {
	return &queueReportClient{
		topicName:      topicName,
		producerClient: producerClient,
	}
}

func (s *queueReportClient) SendReport(ctx context.Context, report JobReport) error {
	return nil
}
