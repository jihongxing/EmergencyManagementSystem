package identity

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func memberAdmin() Member {
	return Member{
		ID: "mem_admin", UserID: "usr_admin", OrganizationID: "org_1",
		OrganizationKind: "enterprise", LoginID: "admin", Status: "active",
		Active: true, Roles: []string{"enterprise_admin"},
	}
}

func memberExecutor() Member {
	return Member{
		ID: "mem_exec", UserID: "usr_exec", OrganizationID: "org_1",
		OrganizationKind: "enterprise", LoginID: "exec", Status: "active",
		Active: true, Roles: []string{"enterprise_executor"},
	}
}

func firstAdmin() Member {
	member := memberAdmin()
	member.FirstAdmin = true
	return member
}

func TestMemberServiceEnforcesOrganizationAndRoleScope(t *testing.T) {
	store := NewMemoryMemberStore(memberAdmin(), memberExecutor())
	service := NewMemberService(store)

	if _, err := service.List(context.Background(), memberExecutor(), "org_1"); !errors.Is(err, ErrAuthorizationDenied) {
		t.Fatalf("executor list got %v", err)
	}
	crossOrg := memberAdmin()
	crossOrg.OrganizationID = "org_2"
	if _, err := service.List(context.Background(), memberAdmin(), "org_2"); !errors.Is(err, ErrAuthorizationDenied) {
		t.Fatalf("cross organization list got %v", err)
	}
	if _, err := service.Create(context.Background(), memberAdmin(), "org_1", MemberCreateRequest{
		LoginID: "inspector", Roles: []string{"inspector"},
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("cross-kind role create got %v", err)
	}
}

func TestMemberCreateStartsPendingAndWritesAudit(t *testing.T) {
	store := NewMemoryMemberStore(memberAdmin())
	service := NewMemberService(store)
	created, err := service.Create(context.Background(), memberAdmin(), "org_1", MemberCreateRequest{
		LoginID: "new-executor", Roles: []string{"enterprise_executor"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != "pending" || created.Active {
		t.Fatalf("created member state: %+v", created)
	}
	if len(store.audits) != 1 || store.audits[0].Action != "member_created" || store.audits[0].ActorMemberID != "mem_admin" {
		t.Fatalf("unexpected create audit: %+v", store.audits)
	}
}

func TestFirstAdminCannotBeUpdated(t *testing.T) {
	store := NewMemoryMemberStore(firstAdmin())
	service := NewMemberService(store)
	status := "revoked"
	if _, err := service.Update(context.Background(), memberAdmin(), "org_1", "mem_admin", MemberUpdateRequest{Status: &status}); !errors.Is(err, ErrMemberProtected) {
		t.Fatalf("first admin update got %v", err)
	}
}

func TestMemberUpdateAuditsRoleAndStatusChanges(t *testing.T) {
	target := memberExecutor()
	store := NewMemoryMemberStore(memberAdmin(), target)
	service := NewMemberService(store)
	roles := []string{"enterprise_admin", "enterprise_executor"}
	status := "suspended"
	updated, err := service.Update(context.Background(), memberAdmin(), "org_1", target.ID, MemberUpdateRequest{Roles: &roles, Status: &status})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != status || len(updated.Roles) != 2 || updated.Active {
		t.Fatalf("unexpected updated member: %+v", updated)
	}
	if len(store.audits) != 1 || store.audits[0].Action != "member_updated" || store.audits[0].Before.Status != "active" || store.audits[0].After.Status != "suspended" {
		t.Fatalf("unexpected update audit: %+v", store.audits)
	}
}

func TestConcurrentCreatesKeepUniqueLogin(t *testing.T) {
	store := NewMemoryMemberStore(memberAdmin())
	service := NewMemberService(store)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.Create(context.Background(), memberAdmin(), "org_1", MemberCreateRequest{
				LoginID: "same-login", Roles: []string{"enterprise_executor"},
			})
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	var success, conflict int
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, ErrMemberConflict) {
			conflict++
		} else {
			t.Fatalf("unexpected concurrent create error: %v", err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
}
