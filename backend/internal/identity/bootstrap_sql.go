package identity

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type SQLBootstrapStore struct {
	db *sql.DB
}

func NewSQLBootstrapStore(db *sql.DB) *SQLBootstrapStore {
	return &SQLBootstrapStore{db: db}
}

func (s *SQLBootstrapStore) Bootstrap(ctx context.Context, actor ControlledActor, request BootstrapRequest) (BootstrapResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BootstrapResult{}, err
	}
	defer tx.Rollback()

	organizationID, err := newIdentityID("org")
	if err != nil {
		return BootstrapResult{}, err
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO ems.organizations (id, kind, name, external_key) VALUES ($1, $2, $3, $4)`,
		organizationID, request.OrganizationKind, request.OrganizationName, request.ExternalKey,
	); err != nil {
		if isConstraintConflict(err) {
			return BootstrapResult{}, ErrOrganizationConflict
		}
		return BootstrapResult{}, err
	}

	member, err := insertFirstAdmin(ctx, tx, organizationID, request.OrganizationKind, request.FirstAdminLogin)
	if err != nil {
		return BootstrapResult{}, err
	}
	if err := insertMaterials(ctx, tx, organizationID, actor.ID, request.Materials); err != nil {
		return BootstrapResult{}, err
	}
	if err := insertAudit(ctx, tx, organizationID, actor.ID, "organization_bootstrap", member.ID); err != nil {
		return BootstrapResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return BootstrapResult{}, err
	}
	return BootstrapResult{
		Organization: Organization{
			ID: organizationID, Kind: request.OrganizationKind, Name: request.OrganizationName,
			ExternalKey: request.ExternalKey, Status: "active",
		},
		FirstAdmin: member,
	}, nil
}

func (s *SQLBootstrapStore) ReplaceFirstAdmin(ctx context.Context, actor ControlledActor, request ReplaceFirstAdminRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var kind string
	if err := tx.QueryRowContext(ctx,
		`SELECT kind FROM ems.organizations WHERE id = $1 FOR UPDATE`, request.OrganizationID,
	).Scan(&kind); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrOrganizationNotFound
		}
		return err
	}
	var oldMemberID string
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM ems.members WHERE organization_id = $1 AND is_first_admin AND status <> 'revoked' FOR UPDATE`,
		request.OrganizationID,
	).Scan(&oldMemberID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrOrganizationNotFound
		}
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE ems.members SET status = 'revoked', updated_at = now() WHERE id = $1`, oldMemberID); err != nil {
		return err
	}
	member, err := insertFirstAdmin(ctx, tx, request.OrganizationID, kind, request.NewAdminLogin)
	if err != nil {
		if isConstraintConflict(err) {
			return ErrOrganizationConflict
		}
		return err
	}
	if err := insertMaterials(ctx, tx, request.OrganizationID, actor.ID, request.Materials); err != nil {
		return err
	}
	if err := insertAudit(ctx, tx, request.OrganizationID, actor.ID, "first_admin_replaced", member.ID); err != nil {
		return err
	}
	return tx.Commit()
}

func insertFirstAdmin(ctx context.Context, tx *sql.Tx, organizationID, kind, loginID string) (Member, error) {
	memberID, err := newIdentityID("mem")
	if err != nil {
		return Member{}, err
	}
	userID, err := newIdentityID("usr")
	if err != nil {
		return Member{}, err
	}
	role := "enterprise_admin"
	if kind == "department" {
		role = "department_admin"
	}
	roles, _ := json.Marshal([]string{role})
	_, err = tx.ExecContext(ctx,
		`INSERT INTO ems.members (id, user_id, organization_id, login_id, status, roles, is_first_admin)
		 VALUES ($1, $2, $3, $4, 'pending', $5, true)`,
		memberID, userID, organizationID, loginID, roles,
	)
	if err != nil {
		return Member{}, err
	}
	return Member{
		ID: memberID, UserID: userID, OrganizationID: organizationID, OrganizationKind: kind,
		LoginID: loginID, Status: "pending", Roles: []string{role}, FirstAdmin: true,
	}, nil
}

func insertMaterials(ctx context.Context, tx *sql.Tx, organizationID, actorID string, materials []Material) error {
	for _, material := range materials {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO ems.organization_materials (organization_id, source_id, verified, verified_at, verified_by)
			 VALUES ($1, $2, $3, now(), $4)`,
			organizationID, material.SourceID, material.Verified, actorID,
		); err != nil {
			return err
		}
	}
	return nil
}

func insertAudit(ctx context.Context, tx *sql.Tx, organizationID, actorID, action, memberID string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO ems.identity_audit_events (organization_id, actor_id, action, member_id, details)
		 VALUES ($1, $2, $3, $4, '{}'::jsonb)`,
		organizationID, actorID, action, memberID,
	)
	return err
}

func newIdentityID(prefix string) (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate %s id: %w", prefix, err)
	}
	return prefix + "_" + hex.EncodeToString(buf), nil
}

func isConstraintConflict(err error) bool {
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "duplicate") || strings.Contains(text, "unique")
}
