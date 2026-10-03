package identity

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type SQLMemberStore struct {
	db *sql.DB
}

func NewSQLMemberStore(db *sql.DB) *SQLMemberStore {
	return &SQLMemberStore{db: db}
}

func scanMember(scanner interface{ Scan(...any) error }) (Member, error) {
	var member Member
	var roles []byte
	if err := scanner.Scan(
		&member.ID, &member.UserID, &member.OrganizationID, &member.OrganizationKind,
		&member.LoginID, &member.Status, &roles, &member.FirstAdmin,
	); err != nil {
		return Member{}, err
	}
	if err := json.Unmarshal(roles, &member.Roles); err != nil {
		return Member{}, err
	}
	member.Active = member.Status == "active"
	return member, nil
}

func (s *SQLMemberStore) ListMembers(ctx context.Context, organizationID string) ([]Member, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT m.id, m.user_id, m.organization_id, o.kind, m.login_id, m.status,
		       m.roles, m.is_first_admin
		FROM ems.members m
		JOIN ems.organizations o ON o.id = m.organization_id
		WHERE m.organization_id = $1
		ORDER BY m.created_at, m.id`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Member, 0)
	for rows.Next() {
		member, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *SQLMemberStore) CreateMember(ctx context.Context, actor Member, member Member) (Member, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Member{}, err
	}
	defer tx.Rollback()
	id, err := newIdentityID("mem")
	if err != nil {
		return Member{}, err
	}
	userID, err := newIdentityID("usr")
	if err != nil {
		return Member{}, err
	}
	roles, err := json.Marshal(member.Roles)
	if err != nil {
		return Member{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO ems.members (id, user_id, organization_id, login_id, status, roles, is_first_admin)
		VALUES ($1, $2, $3, $4, 'pending', $5, false)`,
		id, userID, member.OrganizationID, member.LoginID, roles); err != nil {
		if isConstraintConflict(err) {
			return Member{}, ErrMemberConflict
		}
		return Member{}, err
	}
	created, err := scanMember(tx.QueryRowContext(ctx, `
		SELECT m.id, m.user_id, m.organization_id, o.kind, m.login_id, m.status,
		       m.roles, m.is_first_admin
		FROM ems.members m JOIN ems.organizations o ON o.id = m.organization_id
		WHERE m.id = $1`, id))
	if err != nil {
		return Member{}, err
	}
	if err := insertMemberAudit(ctx, tx, actor, created, "member_created", Member{}, created); err != nil {
		return Member{}, err
	}
	if err := tx.Commit(); err != nil {
		return Member{}, err
	}
	return created, nil
}

func (s *SQLMemberStore) UpdateMember(ctx context.Context, actor Member, organizationID, memberID string, request MemberUpdateRequest) (Member, MemberAudit, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Member{}, MemberAudit{}, err
	}
	defer tx.Rollback()
	before, err := scanMember(tx.QueryRowContext(ctx, `
		SELECT m.id, m.user_id, m.organization_id, o.kind, m.login_id, m.status,
		       m.roles, m.is_first_admin
		FROM ems.members m JOIN ems.organizations o ON o.id = m.organization_id
		WHERE m.id = $1 AND m.organization_id = $2
		FOR UPDATE`, memberID, organizationID))
	if errors.Is(err, sql.ErrNoRows) {
		return Member{}, MemberAudit{}, ErrAuthorizationDenied
	}
	if err != nil {
		return Member{}, MemberAudit{}, err
	}
	if before.FirstAdmin {
		return Member{}, MemberAudit{}, ErrMemberProtected
	}
	if request.Roles != nil {
		if !validRoles(before.OrganizationKind, *request.Roles) {
			return Member{}, MemberAudit{}, ErrInvalidRequest
		}
	}
	roles := before.Roles
	status := before.Status
	if request.Roles != nil {
		roles = *request.Roles
	}
	if request.Status != nil {
		status = *request.Status
	}
	if status != "pending" && status != "active" && status != "suspended" && status != "revoked" {
		return Member{}, MemberAudit{}, ErrInvalidRequest
	}
	encodedRoles, err := json.Marshal(roles)
	if err != nil {
		return Member{}, MemberAudit{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE ems.members SET roles = $2, status = $3, updated_at = now()
		WHERE id = $1`, memberID, encodedRoles, status); err != nil {
		return Member{}, MemberAudit{}, err
	}
	after := before
	after.Roles, after.Status, after.Active = append([]string(nil), roles...), status, status == "active"
	action := "member_updated"
	if request.Roles != nil && request.Status == nil {
		action = "member_roles_updated"
	} else if request.Roles == nil && request.Status != nil {
		action = "member_status_updated"
	}
	audit := MemberAudit{
		OrganizationID: organizationID, ActorMemberID: actor.ID, TargetMemberID: memberID,
		Action: action, Before: before, After: after, OccurredAt: time.Now().UTC(),
	}
	if err := insertMemberAudit(ctx, tx, actor, after, action, before, after); err != nil {
		return Member{}, MemberAudit{}, err
	}
	if err := tx.Commit(); err != nil {
		return Member{}, MemberAudit{}, err
	}
	return after, audit, nil
}

func insertMemberAudit(ctx context.Context, tx *sql.Tx, actor Member, target Member, action string, before, after Member) error {
	beforeJSON, err := json.Marshal(struct {
		ID     string   `json:"id,omitempty"`
		Status string   `json:"status,omitempty"`
		Roles  []string `json:"roles,omitempty"`
	}{before.ID, before.Status, before.Roles})
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(struct {
		ID     string   `json:"id,omitempty"`
		Status string   `json:"status,omitempty"`
		Roles  []string `json:"roles,omitempty"`
	}{after.ID, after.Status, after.Roles})
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO ems.member_audit_events
		  (organization_id, actor_member_id, target_member_id, action, before_state, after_state)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		target.OrganizationID, actor.ID, target.ID, action, beforeJSON, afterJSON)
	return err
}

var _ MemberStore = (*SQLMemberStore)(nil)
