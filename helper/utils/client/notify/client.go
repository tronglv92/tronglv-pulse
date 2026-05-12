package notify

import (
	"context"
	"fmt"

	"pulse/helper/utils/authenticator"
	"pulse/helper/utils/httpc"
)

type Client interface {
	SendNotification(ctx context.Context, payload NotificationReq) error
	SendBulkNotification(ctx context.Context, payload BulkNotificationReq) error
	SendTargetedBulkNotification(ctx context.Context, payload TargetedBulkNotificationReq) error
	SendMail(ctx context.Context, payload EmailReq) error
}

type (
	Config interface {
		GetUrl() string
		GetClientId() string
		GetClientSecret() string
	}
	notifSvc struct {
		httpClient   httpc.Service
		gatewayUrl   string
		clientId     string
		clientSecret string
	}
)

func New(c Config, httpClient httpc.Service) Client {
	return &notifSvc{
		gatewayUrl:   c.GetUrl(),
		clientId:     c.GetClientId(),
		clientSecret: c.GetClientSecret(),
		httpClient:   httpClient,
	}
}

func (s *notifSvc) getUrl(path string) string {
	return fmt.Sprintf("%s/notification-svc/api/v1/%s", s.gatewayUrl, path)
}

func (s *notifSvc) getToken(ctx context.Context) (authenticator.OAuthToken, error) {
	cli := authenticator.New(&authenticator.DefaultConfig{
		Url:        s.gatewayUrl,
		HttpClient: s.httpClient,
	})
	resp, err := cli.SignInWithService(ctx, s.clientId, s.clientSecret)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *notifSvc) SendNotification(ctx context.Context, payload NotificationReq) error {
	t, e := s.getToken(ctx)
	if e != nil {
		return e
	}
	_, err := s.httpClient.Post(ctx,
		s.getUrl("dispatch/notify"),
		payload,
		httpc.WithAuthToken(t.GetAccessToken()),
	)
	return err
}

func (s *notifSvc) SendBulkNotification(ctx context.Context, payload BulkNotificationReq) error {
	t, e := s.getToken(ctx)
	if e != nil {
		return e
	}

	_, err := s.httpClient.Post(ctx,
		s.getUrl("notifications/bulk-create"),
		payload,
		httpc.WithAuthToken(t.GetAccessToken()),
	)
	return err
}

func (s *notifSvc) SendTargetedBulkNotification(ctx context.Context, payload TargetedBulkNotificationReq) error {
	t, e := s.getToken(ctx)
	if e != nil {
		return e
	}

	_, err := s.httpClient.Post(ctx,
		s.getUrl("notifications/bulk-targets"),
		payload,
		httpc.WithAuthToken(t.GetAccessToken()),
	)
	return err
}

func (s *notifSvc) SendMail(ctx context.Context, payload EmailReq) error {
	t, e := s.getToken(ctx)
	if e != nil {
		return e
	}

	_, err := s.httpClient.Post(
		ctx,
		s.getUrl("dispatch/email"),
		payload,
		httpc.WithAuthToken(t.GetAccessToken()),
	)
	return err
}
