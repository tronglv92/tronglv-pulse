package scheduler

import (
	"context"
	"pulse/helper/utils/httpc"
)

type httpReportClient struct {
	domain     string
	httpClient httpc.Service
}

func NewHttpReportClient(httpClient httpc.Service, domain string) JobReporter {
	return &httpReportClient{
		domain:     domain,
		httpClient: httpClient,
	}
}

func (s *httpReportClient) SendReport(ctx context.Context, report JobReport) error {
	return nil
}
