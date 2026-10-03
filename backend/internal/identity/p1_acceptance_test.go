package identity

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func acceptanceMember(id, organizationID, kind, loginID string, roles ...string) Member {
	return Member{
		ID:               id,
		UserID:           "usr_" + id,
		OrganizationID:   organizationID,
		OrganizationKind: kind,
		LoginID:          loginID,
		Status:           "active",
		Active:           true,
		Roles:            roles,
	}
}

func TestP1MinimumOrganizationAcceptance(t *testing.T) {
	ctx := context.Background()
	enterpriseA := acceptanceMember(
		"mem_enterprise_a_admin", "org_enterprise_a", "enterprise", "enterprise-a-admin",
		"enterprise_admin", "enterprise_executor",
	)
	enterpriseB := acceptanceMember(
		"mem_enterprise_b_admin", "org_enterprise_b", "enterprise", "enterprise-b-admin",
		"enterprise_admin",
	)
	department := acceptanceMember(
		"mem_department_admin", "org_department", "department", "department-admin",
		"department_admin", "inspector",
	)

	memberStore := NewMemoryMemberStore(enterpriseA, enterpriseB, department)
	memberService := NewMemberService(memberStore)

	t.Run("each organization is isolated", func(t *testing.T) {
		members, err := memberService.List(ctx, enterpriseA, enterpriseA.OrganizationID)
		if err != nil {
			t.Fatal(err)
		}
		if len(members) != 1 || members[0].OrganizationID != enterpriseA.OrganizationID {
			t.Fatalf("enterprise A list leaked or omitted members: %+v", members)
		}

		if _, err := memberService.List(ctx, enterpriseA, enterpriseB.OrganizationID); !errors.Is(err, ErrAuthorizationDenied) {
			t.Fatalf("enterprise A cross-organization list got %v", err)
		}
		if _, err := memberService.List(ctx, department, enterpriseA.OrganizationID); !errors.Is(err, ErrAuthorizationDenied) {
			t.Fatalf("department cross-organization list got %v", err)
		}
		if _, err := memberService.List(ctx, enterpriseB, enterpriseA.OrganizationID); !errors.Is(err, ErrAuthorizationDenied) {
			t.Fatalf("enterprise B cross-organization list got %v", err)
		}
	})

	t.Run("administrator may also execute within own organization", func(t *testing.T) {
		if got := CanManageMembers(enterpriseA, enterpriseA.OrganizationID); got != "allowed" {
			t.Fatalf("enterprise administrator/executor got %q", got)
		}
		if got := CanManageMembers(department, department.OrganizationID); got != "allowed" {
			t.Fatalf("department administrator/inspector got %q", got)
		}

		created, err := memberService.Create(ctx, enterpriseA, enterpriseA.OrganizationID, MemberCreateRequest{
			LoginID: "enterprise-a-executor",
			Roles:   []string{"enterprise_executor"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if created.OrganizationID != enterpriseA.OrganizationID || created.Status != "pending" || created.Active {
			t.Fatalf("created member escaped lifecycle/scope: %+v", created)
		}
	})

	t.Run("audit evidence records member changes", func(t *testing.T) {
		target := acceptanceMember(
			"mem_a_executor", enterpriseA.OrganizationID, "enterprise", "enterprise-a-executor-2",
			"enterprise_executor",
		)
		store := NewMemoryMemberStore(enterpriseA, target)
		service := NewMemberService(store)

		status := "suspended"
		if _, err := service.Update(ctx, enterpriseA, enterpriseA.OrganizationID, target.ID, MemberUpdateRequest{
			Status: &status,
		}); err != nil {
			t.Fatal(err)
		}
		if len(store.audits) != 1 {
			t.Fatalf("expected one audit event, got %d", len(store.audits))
		}
		audit := store.audits[0]
		if audit.OrganizationID != enterpriseA.OrganizationID ||
			audit.ActorMemberID != enterpriseA.ID ||
			audit.TargetMemberID != target.ID ||
			audit.Action != "member_status_updated" ||
			audit.Before.Status != "active" ||
			audit.After.Status != "suspended" {
			t.Fatalf("unexpected audit evidence: %+v", audit)
		}
	})

	t.Run("revocation takes effect for an existing session", func(t *testing.T) {
		authStore := NewMemoryAuthStore()
		passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct horse"), bcrypt.MinCost)
		if err != nil {
			t.Fatal(err)
		}
		authMember := AuthMember{
			Member:             enterpriseA,
			OrganizationStatus: "active",
			PasswordHash:       string(passwordHash),
		}
		authStore.AddMember(authMember)
		auth := NewAuthService(authStore, func() time.Time {
			return time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
		})

		session, err := auth.Login(ctx, LoginRequest{
			LoginID:  enterpriseA.LoginID,
			Password: "correct horse",
			Client:   ClientFlutterAndroid,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := auth.ValidateAccess(ctx, session.AccessToken); err != nil {
			t.Fatal(err)
		}

		authMember.Status = "revoked"
		authMember.Active = false
		authStore.AddMember(authMember)
		if _, err := auth.ValidateAccess(ctx, session.AccessToken); !errors.Is(err, ErrMemberInactive) {
			t.Fatalf("revoked session got %v, want member inactive", err)
		}
	})
}
