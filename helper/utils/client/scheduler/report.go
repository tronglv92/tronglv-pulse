package scheduler

import (
	"context"
	"time"
)

type TriggerMode string

const (
	TriggerModeAuto   TriggerMode = "auto"   // Scheduled/cron trigger
	TriggerModeManual TriggerMode = "manual" // Manually triggered by user
	TriggerModeAPI    TriggerMode = "api"    // Triggered via API call
	TriggerModeEvent  TriggerMode = "event"  // Triggered by event/webhook
)

type (
	JobReport struct {
		ServiceName     string
		JobName         string
		Error           error
		StartTime       time.Time
		EndTime         time.Time
		TriggerMode     TriggerMode
		TriggeredByUid  string
		TriggeredByName string
	}

	ReportPayload struct {
		ServiceName     string `json:"service_name"`
		JobName         string `json:"job_name"`
		TriggerMode     string `json:"trigger_mode"`
		Status          string `json:"status"`
		Error           string `json:"error,omitempty"`
		StartTime       string `json:"start_time"`
		EndTime         string `json:"end_time"`
		Duration        int64  `json:"duration_ms"`
		TriggeredByUid  string `json:"triggered_by_uid,omitempty"`
		TriggeredByName string `json:"triggered_by_name,omitempty"`
	}
)

type JobReporter interface {
	SendReport(ctx context.Context, report JobReport) error
}
