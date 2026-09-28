package identity

import "strings"

// Member must be loaded from current server-side state, never client claims.
type Member struct {
	UserID           string   `json:"userId"`
	OrganizationID   string   `json:"organizationId"`
	OrganizationKind string   `json:"organizationKind"`
	Active           bool     `json:"active"`
	Roles            []string `json:"roles"`
}

// CanManageMembers does not authorize first-administrator replacement or
// confer administrative inspection qualifications.
func CanManageMembers(member Member, targetOrganization string) string {
	if strings.TrimSpace(member.UserID) == "" || strings.TrimSpace(member.OrganizationID) == "" || len(member.Roles) == 0 {
		return "invalid_identity"
	}
	var admin, executor string
	switch member.OrganizationKind {
	case "enterprise":
		admin, executor = "enterprise_admin", "enterprise_executor"
	case "department":
		admin, executor = "department_admin", "inspector"
	default:
		return "invalid_identity"
	}
	seen := make(map[string]bool)
	for _, role := range member.Roles {
		if seen[role] || (role != admin && role != executor) {
			return "invalid_identity"
		}
		seen[role] = true
	}
	if !member.Active {
		return "inactive"
	}
	if targetOrganization != member.OrganizationID {
		return "scope_denied"
	}
	if !seen[admin] {
		return "role_denied"
	}
	return "allowed"
}
