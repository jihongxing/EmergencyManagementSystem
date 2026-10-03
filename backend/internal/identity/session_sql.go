package identity

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type SQLAuthStore struct {
	db *sql.DB
}

func NewSQLAuthStore(db *sql.DB) *SQLAuthStore {
	return &SQLAuthStore{db: db}
}

func (s *SQLAuthStore) FindMemberByLogin(ctx context.Context, loginID string) (AuthMember, error) {
	var member AuthMember
	var roles []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT m.id, m.user_id, m.organization_id, o.kind, m.login_id, m.status,
		       m.roles, m.is_first_admin, o.status, COALESCE(m.password_hash, '')
		FROM ems.members m
		JOIN ems.organizations o ON o.id = m.organization_id
		WHERE m.login_id = $1`, loginID).Scan(
		&member.ID, &member.UserID, &member.OrganizationID, &member.OrganizationKind,
		&member.LoginID, &member.Status, &roles, &member.FirstAdmin,
		&member.OrganizationStatus, &member.PasswordHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return AuthMember{}, ErrAuthenticationFailed
	}
	if err != nil {
		return AuthMember{}, err
	}
	if err := json.Unmarshal(roles, &member.Roles); err != nil {
		return AuthMember{}, err
	}
	member.Active = member.Status == "active"
	return member, nil
}

func (s *SQLAuthStore) SetPasswordAndActivate(ctx context.Context, memberID, hash string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE ems.members
		SET password_hash = $2, status = 'active', updated_at = now()
		WHERE id = $1 AND status = 'pending'`, memberID, hash)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return ErrMemberInactive
	}
	return nil
}

func (s *SQLAuthStore) CreateSession(ctx context.Context, session Session) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO ems.sessions
		  (id, member_id, client, access_token_hash, refresh_token_hash, access_expires_at, refresh_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		session.ID, session.MemberID, session.Client, tokenHash(session.AccessToken),
		tokenHash(session.RefreshToken), session.AccessExpiresAt, session.RefreshExpiresAt)
	return err
}

func (s *SQLAuthStore) sessionQuery(ctx context.Context, predicate string, value string) (SessionState, error) {
	var state SessionState
	var roles []byte
	var revoked sql.NullTime
	err := s.db.QueryRowContext(ctx, `
		SELECT s.id, s.member_id, s.client, s.access_expires_at, s.refresh_expires_at, s.revoked_at,
		       m.id, m.user_id, m.organization_id, o.kind, m.login_id, m.status,
		       m.roles, m.is_first_admin, o.status
		FROM ems.sessions s
		JOIN ems.members m ON m.id = s.member_id
		JOIN ems.organizations o ON o.id = m.organization_id
		WHERE `+predicate, value).Scan(
		&state.ID, &state.MemberID, &state.Client, &state.AccessExpiresAt, &state.RefreshExpiresAt, &revoked,
		&state.Member.ID, &state.Member.UserID, &state.Member.OrganizationID, &state.Member.OrganizationKind,
		&state.Member.LoginID, &state.Member.Status, &roles, &state.Member.FirstAdmin,
		&state.OrganizationStatus,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return SessionState{}, ErrTokenInvalid
	}
	if err != nil {
		return SessionState{}, err
	}
	if revoked.Valid {
		state.RevokedAt = &revoked.Time
	}
	if err := json.Unmarshal(roles, &state.Member.Roles); err != nil {
		return SessionState{}, err
	}
	state.Member.Active = state.Member.Status == "active"
	return state, nil
}

func (s *SQLAuthStore) FindSessionByAccessHash(ctx context.Context, hash string) (SessionState, error) {
	return s.sessionQuery(ctx, "s.access_token_hash = $1", hash)
}

func (s *SQLAuthStore) FindSessionByRefreshHash(ctx context.Context, hash string) (SessionState, error) {
	return s.sessionQuery(ctx, "s.refresh_token_hash = $1", hash)
}

func (s *SQLAuthStore) RotateRefresh(ctx context.Context, oldID string, next Session, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var revoked sql.NullTime
	var expires time.Time
	if err := tx.QueryRowContext(ctx,
		`SELECT revoked_at, refresh_expires_at FROM ems.sessions WHERE id = $1 FOR UPDATE`,
		oldID).Scan(&revoked, &expires); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTokenReplayed
		}
		return err
	}
	if revoked.Valid || !expires.After(now) {
		return ErrTokenReplayed
	}
	if _, err := tx.ExecContext(ctx, `UPDATE ems.sessions SET revoked_at = now() WHERE id = $1`, oldID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO ems.sessions
		  (id, member_id, client, access_token_hash, refresh_token_hash, access_expires_at, refresh_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		next.ID, next.MemberID, next.Client, tokenHash(next.AccessToken),
		tokenHash(next.RefreshToken), next.AccessExpiresAt, next.RefreshExpiresAt); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLAuthStore) RevokeSession(ctx context.Context, accessHash string, when time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE ems.sessions SET revoked_at = $2 WHERE access_token_hash = $1 AND revoked_at IS NULL`, accessHash, when)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return ErrAuthenticationRequired
	}
	return nil
}

func (s *SQLAuthStore) CreatePasswordReset(ctx context.Context, memberID, hash string, expires time.Time) error {
	id, err := newIdentityID("rst")
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO ems.password_resets (id, member_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)`, id, memberID, hash, expires)
	return err
}

func (s *SQLAuthStore) ConsumePasswordResetAndSetPassword(ctx context.Context, hash, passwordHash string, when time.Time) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var resetID, memberID string
	if err := tx.QueryRowContext(ctx, `
		SELECT id, member_id FROM ems.password_resets
		WHERE token_hash = $1 AND used_at IS NULL AND revoked_at IS NULL AND expires_at > $2
		FOR UPDATE`, hash, when).Scan(&resetID, &memberID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrTokenInvalid
		}
		return "", err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE ems.password_resets SET used_at = $2 WHERE id = $1`, resetID, when); err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE ems.members SET password_hash = $2, status = 'active', updated_at = now()
		WHERE id = $1`, memberID, passwordHash); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return memberID, nil
}

func (s *SQLAuthStore) RevokeMemberSessions(ctx context.Context, memberID string, when time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE ems.sessions SET revoked_at = $2 WHERE member_id = $1 AND revoked_at IS NULL`, memberID, when)
	return err
}

var _ AuthStore = (*SQLAuthStore)(nil)
