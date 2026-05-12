package notify

import "time"

type NotificationReq struct {
	NotificationDataReq
	TargetId string `json:"target_id"`
}

type TargetedBulkNotificationReq struct {
	NotificationDataReq
	TargetIds []string `json:"target_ids"`
}

type BulkNotificationReq struct {
	Items []NotificationReq `json:"items"`
}

type NotificationDataReq struct {
	TraceId    string       `json:"trace_id"`
	TargetType int          `json:"target_type,optional"`
	TitleVn    string       `json:"title"`
	ContentVn  string       `json:"content"`
	TitleEn    string       `json:"title_en,optional"`
	ContentEn  string       `json:"content_en,optional"`
	Channels   []string     `json:"channels"`
	Platforms  []int        `json:"platforms,optional"`
	Summary    string       `json:"summary,optional"`
	Url        string       `json:"url,optional"`
	ImageUrl   string       `json:"image_url,optional"`
	IconUrl    string       `json:"icon_url,optional"`
	Sound      string       `json:"sound,optional"`
	Attributes NotifAttrReq `json:"attributes,optional"`
	MarkedRead bool         `json:"marked_read,optional"`
	PushedAt   *time.Time   `json:"pushed_at,optional"`
}

type NotifAttrReq struct {
	ActionId    int    `json:"action_id,optional"`
	ActionType  int    `json:"action_type,optional"`
	ActionUrl   string `json:"action_url,optional"`
	OrderNumber string `json:"order_number,optional"`
}

type EmailReq struct {
	TraceId   string   `json:"trace_id"`
	Account   string   `json:"account"`
	Subject   string   `json:"subject"`
	Sender    string   `json:"sender"`
	Recipient string   `json:"recipient"`
	CC        []string `json:"cc"`
	Bcc       []string `json:"bcc"`
	Body      string   `json:"html_body,optional"`
}
