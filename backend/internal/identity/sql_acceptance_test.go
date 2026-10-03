package identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"emergency-management/backend/internal/database"
)

func TestP1SQLMinimumOrganizationAcceptance(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL required for SQL identity acceptance")
	}

	db, err := database.Open(url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var migrationVersion int64
	if err := db.QueryRowContext(ctx,
		`SELECT version_id FROM public.goose_db_version ORDER BY id DESC LIMIT 1`,
	).Scan(&migrationVersion); err != nil {
		t.Fatal(err)
	}
	if migrationVersion < database.CurrentMigrationVersion {
		t.Fatalf("database migration version %d is below %d", migrationVersion, database.CurrentMigrationVersion)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	enterpriseAKey := "p1-sql-enterprise-a-" + suffix
	bootstrap := NewBootstrapService(NewSQLBootstrapStore(db))
	actor := ControlledActor{ID: "platform-acceptance-actor"}
	materials := []Material{{SourceID: "acceptance-material", Verified: true}}

	enterpriseA, err := bootstrap.Bootstrap(ctx, actor, BootstrapRequest{
		OrganizationKind: "enterprise",
		OrganizationName: "P1 Enterprise A",
		ExternalKey:      enterpriseAKey,
		FirstAdminLogin:  "p1-sql-enterprise-a-admin-" + suffix,
		Materials:        materials,
	})
	if err != nil {
		t.Fatal(err)
	}
	enterpriseB, err := bootstrap.Bootstrap(ctx, actor, BootstrapRequest{
		OrganizationKind: "enterprise",
		OrganizationName: "P1 Enterprise B",
		ExternalKey:      "p1-sql-enterprise-b-" + suffix,
		FirstAdminLogin:  "p1-sql-enterprise-b-admin-" + suffix,
		Materials:        materials,
	})
	if err != nil {
		t.Fatal(err)
	}
	department, err := bootstrap.Bootstrap(ctx, actor, BootstrapRequest{
		OrganizationKind: "department",
		OrganizationName: "P1 Department",
		ExternalKey:      "p1-sql-department-" + suffix,
		FirstAdminLogin:  "p1-sql-department-admin-" + suffix,
		Materials:        materials,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bootstrap.Bootstrap(ctx, actor, BootstrapRequest{
		OrganizationKind: "enterprise",
		OrganizationName: "P1 Duplicate",
		ExternalKey:      enterpriseAKey,
		FirstAdminLogin:  "p1-sql-duplicate-admin-" + suffix,
		Materials:        materials,
	}); !errors.Is(err, ErrOrganizationConflict) {
		t.Fatalf("duplicate external key returned %v", err)
	}

	auth := NewAuthService(NewSQLAuthStore(db), func() time.Time {
		return time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	})
	if err := auth.Activate(ctx, enterpriseA.FirstAdmin.ID, "correct horse"); err != nil {
		t.Fatal(err)
	}
	if err := auth.Activate(ctx, department.FirstAdmin.ID, "department horse"); err != nil {
		t.Fatal(err)
	}

	memberStore := NewSQLMemberStore(db)
	memberService := NewMemberService(memberStore)
	enterpriseAMembers, err := memberStore.ListMembers(ctx, enterpriseA.Organization.ID)
	if err != nil {
		t.Fatal(err)
	}
	var activeEnterpriseAAdmin Member
	for _, member := range enterpriseAMembers {
		if member.ID == enterpriseA.FirstAdmin.ID {
			activeEnterpriseAAdmin = member
			break
		}
	}
	if activeEnterpriseAAdmin.Status != "active" {
		t.Fatalf("activated administrator was not reloaded as active: %+v", activeEnterpriseAAdmin)
	}

	departmentMembers, err := memberStore.ListMembers(ctx, department.Organization.ID)
	if err != nil {
		t.Fatal(err)
	}
	var activeDepartmentAdmin Member
	for _, member := range departmentMembers {
		if member.ID == department.FirstAdmin.ID {
			activeDepartmentAdmin = member
			break
		}
	}

	created, err := memberService.Create(ctx, activeEnterpriseAAdmin, enterpriseA.Organization.ID, MemberCreateRequest{
		LoginID: "p1-sql-enterprise-a-executor-" + suffix,
		Roles:   []string{"enterprise_executor"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != "pending" || created.Active {
		t.Fatalf("new member lifecycle is %+v", created)
	}

	if _, err := memberService.List(ctx, activeEnterpriseAAdmin, enterpriseB.Organization.ID); !errors.Is(err, ErrAuthorizationDenied) {
		t.Fatalf("cross-organization enterprise access returned %v", err)
	}
	if _, err := memberService.List(ctx, activeDepartmentAdmin, enterpriseA.Organization.ID); !errors.Is(err, ErrAuthorizationDenied) {
		t.Fatalf("department-to-enterprise access returned %v", err)
	}

	status := "suspended"
	if _, err := memberService.Update(ctx, activeEnterpriseAAdmin, enterpriseA.Organization.ID, created.ID, MemberUpdateRequest{
		Status: &status,
	}); err != nil {
		t.Fatal(err)
	}

	var organizationID, actorID, targetID, action string
	var beforeStatus, afterStatus string
	if err := db.QueryRowContext(ctx, `
		SELECT organization_id, actor_member_id, target_member_id, action,
		       before_state->>'status', after_state->>'status'
		FROM ems.member_audit_events
		WHERE target_member_id = $1
		ORDER BY id DESC LIMIT 1`, created.ID,
	).Scan(&organizationID, &actorID, &targetID, &action, &beforeStatus, &afterStatus); err != nil {
		t.Fatal(err)
	}
	if organizationID != enterpriseA.Organization.ID ||
		actorID != enterpriseA.FirstAdmin.ID ||
		targetID != created.ID ||
		action != "member_status_updated" ||
		beforeStatus != "pending" ||
		afterStatus != "suspended" {
		t.Fatalf("unexpected member audit: org=%q actor=%q target=%q action=%q before=%q after=%q",
			organizationID, actorID, targetID, action, beforeStatus, afterStatus)
	}

	session, err := auth.Login(ctx, LoginRequest{
		LoginID:  enterpriseA.FirstAdmin.LoginID,
		Password: "correct horse",
		Client:   ClientFlutterAndroid,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := auth.ValidateAccess(ctx, session.AccessToken); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx,
		`UPDATE ems.members SET status = 'revoked', updated_at = now() WHERE id = $1`,
		enterpriseA.FirstAdmin.ID,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.ValidateAccess(ctx, session.AccessToken); !errors.Is(err, ErrMemberInactive) {
		t.Fatalf("revoked member access returned %v", err)
	}

	if got := CanManageMembers(activeEnterpriseAAdmin, enterpriseA.Organization.ID); got != "allowed" {
		t.Fatalf("enterprise admin/executor compatibility returned %q", got)
	}
	if got := CanManageMembers(activeDepartmentAdmin, department.Organization.ID); got != "allowed" {
		t.Fatalf("department admin/inspector compatibility returned %q", got)
	}

	var passwordHash sql.NullString
	if err := db.QueryRowContext(ctx,
		`SELECT password_hash FROM ems.members WHERE id = $1`,
		enterpriseA.FirstAdmin.ID,
	).Scan(&passwordHash); err != nil {
		t.Fatal(err)
	}
	if !passwordHash.Valid || passwordHash.String == "" {
		t.Fatal("activated administrator password was not persisted")
	}
}
