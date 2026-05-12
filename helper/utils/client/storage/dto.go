package storage

import (
	"pulse/helper/utils/internal/attachment"
	pmcpb "pulse/helper/utils/pmc/protobuf"
	"pulse/helper/utils/toolkit/filex"
	"google.golang.org/protobuf/types/known/durationpb"
	"time"
)

type ActionType int32

const (
	UploadType ActionType = iota
	ExportType
)

type UploadAttachmentRequest struct {
	Acl              *string            `json:"acl,omitempty"`
	Action           *ActionType        `json:"action,omitempty"`
	Duration         *time.Duration     `json:"duration,omitempty"`
	EmployeeId       *string            `json:"employee_id,omitempty"`
	IsOverride       *bool              `json:"is_override,omitempty"`
	ServiceName      string             `json:"service_name"`
	Attachments      []filex.FileEntity `json:"attachments"`
	IsPresigned      *bool              `json:"is_presigned,omitempty"`
	BucketAlias      *string            `json:"bucket_alias,omitempty"`
	OriginalFilename *bool              `json:"original_filename,omitempty"`
}

func (r *UploadAttachmentRequest) GetAcl() *string            { return r.Acl }
func (r *UploadAttachmentRequest) GetEmployeeId() *string     { return r.EmployeeId }
func (r *UploadAttachmentRequest) GetServiceName() string     { return r.ServiceName }
func (r *UploadAttachmentRequest) GetIsOverride() *bool       { return r.IsOverride }
func (r *UploadAttachmentRequest) GetIsPresigned() *bool      { return r.IsPresigned }
func (r *UploadAttachmentRequest) GetBucketAlias() *string    { return r.BucketAlias }
func (r *UploadAttachmentRequest) GetOriginalFilename() *bool { return r.OriginalFilename }
func (r *UploadAttachmentRequest) GetDuration() *durationpb.Duration {
	if r.Duration == nil {
		return nil
	}
	return durationpb.New(*r.Duration)
}
func (r *UploadAttachmentRequest) GetAttachments() []*attachment.FileInfo {
	var attachments []*attachment.FileInfo
	for _, v := range r.Attachments {
		attachments = append(attachments, &attachment.FileInfo{
			FileName: v.GetName(),
			FileData: v.GetData(),
			FileSize: v.GetSize(),
		})
	}
	return attachments
}
func (r *UploadAttachmentRequest) GetAction() *attachment.ActionType {
	if r.Action == nil {
		return nil
	}
	return (*attachment.ActionType)(r.Action)
}

type PreSignedUrlsRequest struct {
	Urls        []string       `json:"urls"`
	IsPut       *bool          `json:"is_put,omitempty"`
	Action      *ActionType    `json:"action,omitempty"`
	Duration    *time.Duration `json:"duration,omitempty"`
	EmployeeId  *string        `json:"employee_id,omitempty"`
	BucketAlias *string        `json:"bucket_alias,omitempty"`
	ServiceName string         `json:"service_name"`
}

func (r *PreSignedUrlsRequest) GetUrls() []string { return r.Urls }
func (r *PreSignedUrlsRequest) GetIsPut() *bool   { return r.IsPut }
func (r *PreSignedUrlsRequest) GetAction() *attachment.ActionType {
	if r.Action == nil {
		return nil
	}
	return (*attachment.ActionType)(r.Action)
}
func (r *PreSignedUrlsRequest) GetDuration() *durationpb.Duration {
	if r.Duration == nil {
		return nil
	}
	return durationpb.New(*r.Duration)
}
func (r *PreSignedUrlsRequest) GetEmployeeId() *string  { return r.EmployeeId }
func (r *PreSignedUrlsRequest) GetBucketAlias() *string { return r.BucketAlias }
func (r *PreSignedUrlsRequest) GetServiceName() string  { return r.ServiceName }

type UploadAttachmentResponse struct {
	Attachments []UploadAttachmentItem `json:"attachments"`
}

func (r *UploadAttachmentResponse) GetAttachments() []UploadAttachmentItem { return r.Attachments }

type UploadAttachmentItem struct {
	Url   string        `json:"url,omitempty"`
	Error *pmcpb.Status `json:"error,omitempty"`
}

func (i *UploadAttachmentItem) IsSuccess() bool         { return i.Error == nil }
func (i *UploadAttachmentItem) GetUrl() string          { return i.Url }
func (i *UploadAttachmentItem) GetError() *pmcpb.Status { return i.Error }

type PreSignedUrlsResponse struct {
	SignedUrls []string `json:"signed_urls"`
}

func (r *PreSignedUrlsResponse) GetSignedUrls() []string { return r.SignedUrls }
