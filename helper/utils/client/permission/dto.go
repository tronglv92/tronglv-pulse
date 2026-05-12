package permission

type GetAccessRequest struct {
	ExternalId         string `json:"external_id" validate:"min=4,max=50"`
	IncludeRoles       bool   `json:"include_roles"`
	IncludeGroups      bool   `json:"include_groups"`
	IncludePermissions bool   `json:"include_permissions"`
}

type GetAccessCodesResponse struct {
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
}

type GetInternalRoleCodesRequest struct {
	ExternalId string `json:"external_id" validate:"min=4,max=50"`
}

type GetInternalRoleCodesResponse struct{}
