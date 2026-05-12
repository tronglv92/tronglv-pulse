package storage

import (
	"context"
	"pulse/helper/utils/grpcx"
	"pulse/helper/utils/internal/attachment"
	"pulse/helper/utils/toolkit/contextx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type Client interface {
	Upload(ctx context.Context, req *UploadAttachmentRequest, opts ...grpc.CallOption) (*UploadAttachmentResponse, error)
	PreSignedUrls(ctx context.Context, req *PreSignedUrlsRequest, opts ...grpc.CallOption) (*PreSignedUrlsResponse, error)
}

type clientImplSvc struct {
	clientManager *grpcx.ClientManager[attachment.AttachmentClient]
}

func NewClient(config zrpc.RpcClientConf) Client {
	return &clientImplSvc{
		clientManager: grpcx.NewClientManager(
			config,
			func(conn *grpc.ClientConn) attachment.AttachmentClient {
				return attachment.NewAttachmentClient(conn)
			},
		),
	}
}

func (s *clientImplSvc) Upload(ctx context.Context, in *UploadAttachmentRequest, opts ...grpc.CallOption) (*UploadAttachmentResponse, error) {
	conn, err := s.clientManager.Get()
	if err != nil {
		return nil, err
	}

	req := &attachment.UploadAttachmentRequest{
		Acl:              in.GetAcl(),
		Action:           in.GetAction(),
		Duration:         in.GetDuration(),
		Attachments:      in.GetAttachments(),
		EmployeeId:       in.GetEmployeeId(),
		ServiceName:      in.GetServiceName(),
		IsOverride:       in.GetIsOverride(),
		IsPresigned:      in.GetIsPresigned(),
		BucketAlias:      in.GetBucketAlias(),
		OriginalFilename: in.GetOriginalFilename(),
	}
	resp, err := conn.Upload(ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	var attachments []UploadAttachmentItem
	for _, val := range resp.Attachments {
		attachments = append(attachments, UploadAttachmentItem{
			Url:   val.GetUrl(),
			Error: val.GetError(),
		})
	}
	return &UploadAttachmentResponse{
		Attachments: attachments,
	}, nil
}

func (s *clientImplSvc) PreSignedUrls(ctx context.Context, in *PreSignedUrlsRequest, opts ...grpc.CallOption) (*PreSignedUrlsResponse, error) {
	conn, err := s.clientManager.Get()
	if err != nil {
		return nil, err
	}

	req := &attachment.PreSignedUrlsRequest{
		Urls:        in.GetUrls(),
		IsPut:       in.GetIsPut(),
		Action:      in.GetAction(),
		Duration:    in.GetDuration(),
		BucketAlias: in.GetBucketAlias(),
		EmployeeId:  in.GetEmployeeId(),
		ServiceName: in.GetServiceName(),
	}
	item, err := conn.PreSignedUrls(contextx.WithAuthorizationFromContext(ctx), req, opts...)
	if err != nil {
		return nil, err
	}
	return &PreSignedUrlsResponse{
		SignedUrls: item.GetSignedUrls(),
	}, nil
}
