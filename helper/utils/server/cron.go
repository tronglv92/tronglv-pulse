package server

import (
	"context"
	"pulse/helper/utils/authenticator"
	"pulse/helper/utils/client/scheduler"
	"pulse/helper/utils/identity"
	"pulse/helper/utils/recovery"
	"pulse/helper/utils/toolkit/timex"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	"time"
)

type (
	CronHandler interface {
		Command() *cobra.Command
	}

	CronServer interface {
		GetCommand() *cobra.Command
		Register(ctx context.Context, clis []CronHandler)
	}

	CronOption func(options *cronOptions)

	cronOptions struct {
		transport     authenticator.AuthTransport
		jobReporter   scheduler.JobReporter
		panicReporter recovery.PanicReporter
	}

	CronJobConf struct {
		service.ServiceConf
		Url          string `json:"url,optional"`
		ClientId     string `json:"client-id,optional"`
		ClientSecret string `json:"client-secret,optional"`
	}

	cronServer struct {
		cmd    *cobra.Command
		config CronJobConf
		*cronOptions
	}
)

func NewCron(c CronJobConf, opts ...CronOption) CronServer {
	c.MustSetUp()
	r := &cronServer{
		config:      c,
		cronOptions: &cronOptions{},
		cmd: &cobra.Command{
			Use:   "cron",
			Args:  cobra.NoArgs,
			Short: "exec cron job",
		},
	}
	for _, opt := range opts {
		opt(r.cronOptions)
	}
	return r
}

func (s *cronServer) GetCommand() *cobra.Command {
	return s.cmd
}

func (s *cronServer) Register(ctx context.Context, clis []CronHandler) {
	tracer := otel.Tracer(trace.TraceName)

	for _, handler := range clis {
		cmd := handler.Command()
		origRunE := cmd.RunE

		cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
			// Start time
			startTime := timex.Now()

			// Start span
			ctxWithSpan, span := tracer.Start(ctx, "CronJob."+cmd.Name())
			defer func() {
				if r := recover(); r != nil && s.panicReporter != nil {

					recovery.RecoverWithReporter(ctxWithSpan, s.panicReporter)

					err = fmt.Errorf("panic recovered: %v", r)
				}

				if e := s.report(ctxWithSpan, cmd.Name(), err, startTime); e != nil {
					logx.WithContext(ctxWithSpan).Error(e)
				}

				span.End()
				trace.StopAgent()
				logx.WithContext(ctxWithSpan).Infof("End Cron Command: %s", cmd.Name())
			}()

			logx.WithContext(ctxWithSpan).Infof("Start Cron Command: %s", cmd.Name())
			cmd.SetContext(ctxWithSpan)

			// Execute job
			err = origRunE(cmd, args)

			return err
		}
		s.cmd.AddCommand(cmd)
	}

	if err := s.cmd.ExecuteContext(ctx); err != nil {
		logx.WithContext(ctx).Errorf("Cron command failed: %v", err)
	}
}

func (s *cronServer) authentication(ctx context.Context, clientId, clientSecret string) (identity.Claims, error) {
	if s.transport == nil {
		return nil, fmt.Errorf("transport is nil")
	}

	auth := authenticator.NewAuthenticator(s.transport)
	resp, err := auth.SignInWithService(ctx, clientId, clientSecret)
	if err != nil {
		return nil, err
	}

	mapClaims, e := auth.VerifyToken(ctx, resp.GetAccessToken())
	if e != nil {
		return nil, err
	}
	return mapClaims, nil
}

func (s *cronServer) report(ctx context.Context, jobName string, err error, startTime time.Time) error {
	if s.jobReporter == nil {
		return nil
	}
	return s.jobReporter.SendReport(ctx, scheduler.JobReport{
		ServiceName:     s.config.Name,
		JobName:         jobName,
		Error:           err,
		StartTime:       startTime,
		EndTime:         timex.Now(),
		TriggerMode:     scheduler.TriggerModeAuto,
		TriggeredByName: s.config.Name,
	})
}

func WithAuthTransport(transport authenticator.AuthTransport) CronOption {
	return func(options *cronOptions) {
		options.transport = transport
	}
}

func WithPanicReporter(reporter recovery.PanicReporter) CronOption {
	return func(options *cronOptions) {
		options.panicReporter = reporter
	}
}

func WithJobReporter(reporter scheduler.JobReporter) CronOption {
	return func(options *cronOptions) {
		options.jobReporter = reporter
	}
}
