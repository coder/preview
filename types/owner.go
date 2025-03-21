package types

import (
	"github.com/google/uuid"
)

// Based on https://github.com/coder/terraform-provider-coder/blob/9a745586b23a9cb5de2f65a2dcac12e48b134ffa/provider/workspace_owner.go#L72
type WorkspaceOwner struct {
	ID              uuid.UUID                `json:"id"`
	Name            string                   `json:"name"`
	FullName        string                   `json:"full_name"`
	Email           string                   `json:"email"`
	SSHPublicKey    string                   `json:"ssh_public_key"`
	SSHPrivateKey   string                   `json:"ssh_private_key" tfsdk:",sensitive"`
	Groups          []string                 `json:"groups"`
	SessionToken    string                   `json:"session_token"`
	OIDCAccessToken string                   `json:"oidc_access_token"`
	LoginType       string                   `json:"login_type"`
	RBACRoles       []WorkspaceOwnerRBACRole `json:"rbac_roles"`
}

type WorkspaceOwnerRBACRole struct {
	Name  string    `json:"name"`
	OrgID uuid.UUID `json:"org_id"`
}

// terraform-provider-framework style
// type UserResourceModel struct {
// 	ID UUID `tfsdk:"id"`

// 	Username  types.String `tfsdk:"username"`
// 	Name      types.String `tfsdk:"name"`
// 	Email     types.String `tfsdk:"email"`
// 	Roles     types.Set    `tfsdk:"roles"`      // owner, template-admin, user-admin, auditor (member is implicit)
// 	LoginType types.String `tfsdk:"login_type"` // none, password, github, oidc
// 	Password  types.String `tfsdk:"password"`   // only when login_type is password
// 	Suspended types.Bool   `tfsdk:"suspended"`
// }
