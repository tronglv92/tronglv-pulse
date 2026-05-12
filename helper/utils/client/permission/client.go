package permission

import (
	"context"
	"pulse/helper/utils/grpcx"
	permissionpb "pulse/helper/utils/internal/permission"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type PermissionClient interface {
	GetAccessCodes(ctx context.Context, req *GetAccessRequest, opts ...grpc.CallOption) (*GetAccessCodesResponse, error)
	GetInternalRoleCodes(ctx context.Context, in *GetInternalRoleCodesRequest, opts ...grpc.CallOption) (*GetAccessCodesResponse, error)
}

type permissionClient struct {
	clientManager *grpcx.ClientManager[permissionpb.SubjectClient]
}

func NewPermissionClient(config zrpc.RpcClientConf) PermissionClient {
	return &permissionClient{
		clientManager: grpcx.NewClientManager(
			config,
			func(conn *grpc.ClientConn) permissionpb.SubjectClient {
				return permissionpb.NewSubjectClient(conn)
			},
		),
	}
}

func (s *permissionClient) GetAccessCodes(ctx context.Context, in *GetAccessRequest, opts ...grpc.CallOption) (*GetAccessCodesResponse, error) {
	conn, err := s.clientManager.Get()
	if err != nil {
		return nil, err
	}

	result, err := conn.GetAccessCodes(ctx,
		&permissionpb.GetAccessRequest{
			ExternalId:         in.ExternalId,
			IncludeRoles:       in.IncludeRoles,
			IncludeGroups:      in.IncludeGroups,
			IncludePermissions: in.IncludePermissions,
		},
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return &GetAccessCodesResponse{
		Permissions: result.GetPermissions(),
		Roles:       result.GetRoles(),
	}, nil
}

func (s *permissionClient) GetInternalRoleCodes(ctx context.Context, in *GetInternalRoleCodesRequest, opts ...grpc.CallOption) (*GetAccessCodesResponse, error) {
	conn, err := s.clientManager.Get()
	if err != nil {
		return nil, err
	}

	result, err := conn.GetInternalRoleCodes(ctx,
		&permissionpb.GetInternalRoleCodesRequest{
			ExternalId: in.ExternalId,
		},
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return &GetAccessCodesResponse{
		Permissions: result.GetPermissions(),
		Roles:       result.GetRoles(),
	}, nil
}
