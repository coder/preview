package types

import (
	"github.com/google/uuid"
)

// Based on https://github.com/coder/terraform-provider-coder/blob/9a745586b23a9cb5de2f65a2dcac12e48b134ffa/provider/workspace_owner.go#L72
type WorkspaceOwner struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	SSHPublicKey string    `json:"ssh_public_key"`
	// SSHPrivateKey is intentionally omitted for now, due to the security risk
	// that exposing it poses.
	// SSHPrivateKey string `json:"ssh_private_key"`
	Groups []string `json:"groups"`
	// SessionToken is intentionally omitted for now, due to the security risk
	// that exposing it poses.
	// SessionToken string `json:"session_token"`
	// OIDCAccessToken is intentionally omitted for now, due to the security risk
	// that exposing it poses.
	// OIDCAccessToken string `json:"oidc_access_token"`
	LoginType string                   `json:"login_type"`
	RBACRoles []WorkspaceOwnerRBACRole `json:"rbac_roles"`
}

type WorkspaceOwnerRBACRole struct {
	Name  string    `json:"name"`
	OrgID uuid.UUID `json:"org_id"`
}
