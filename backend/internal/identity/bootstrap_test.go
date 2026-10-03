package identity

import (
	"context"
	"errors"
	"testing"
)

func TestBootstrapCreatesPendingFirstAdminWithVerifiedMaterials(t *testing.T) {
	store := NewMemoryBootstrapStore()
	result, err := NewBootstrapService(store).Bootstrap(context.Background(), ControlledActor{ID: "impl-1"}, BootstrapRequest{
		OrganizationKind: "enterprise", OrganizationName: "Fixture Unit", ExternalKey: "unit-001",
		FirstAdminLogin: "admin@example.invalid", Materials: []Material{{SourceID: "doc-1", Verified: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.FirstAdmin.Status != "pending" || !result.FirstAdmin.IsFirstAdmin() {
		t.Fatalf("unexpected first admin: %#v", result.FirstAdmin)
	}
	if len(store.organizations) != 1 || len(store.materials) != 1 || len(store.audits) != 1 {
		t.Fatalf("expected organization, material and audit persistence: %#v %#v %#v", store.organizations, store.materials, store.audits)
	}
}

func TestBootstrapRejectsUnverifiedAndDuplicateOrganizations(t *testing.T) {
	store := NewMemoryBootstrapStore()
	service := NewBootstrapService(store)
	_, err := service.Bootstrap(context.Background(), ControlledActor{ID: "impl-1"}, BootstrapRequest{
		OrganizationKind: "enterprise", OrganizationName: "Fixture Unit", ExternalKey: "unit-001",
		FirstAdminLogin: "admin@example.invalid", Materials: []Material{{SourceID: "doc-1", Verified: false}},
	})
	if !errors.Is(err, ErrExternalUnverified) {
		t.Fatalf("got %v, want ErrExternalUnverified", err)
	}
	request := BootstrapRequest{OrganizationKind: "enterprise", OrganizationName: "Fixture Unit", ExternalKey: "unit-001", FirstAdminLogin: "admin@example.invalid", Materials: []Material{{SourceID: "doc-1", Verified: true}}}
	if _, err := service.Bootstrap(context.Background(), ControlledActor{ID: "impl-1"}, request); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Bootstrap(context.Background(), ControlledActor{ID: "impl-1"}, request); !errors.Is(err, ErrOrganizationConflict) {
		t.Fatalf("got %v, want ErrOrganizationConflict", err)
	}
}

func TestReplaceFirstAdminRevokesOldAndAudits(t *testing.T) {
	store := NewMemoryBootstrapStore()
	service := NewBootstrapService(store)
	result, err := service.Bootstrap(context.Background(), ControlledActor{ID: "impl-1"}, BootstrapRequest{
		OrganizationKind: "department", OrganizationName: "Fixture Department", ExternalKey: "dept-001",
		FirstAdminLogin: "old@example.invalid", Materials: []Material{{SourceID: "doc-1", Verified: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = service.ReplaceFirstAdmin(context.Background(), ControlledActor{ID: "impl-2"}, ReplaceFirstAdminRequest{
		OrganizationID: result.Organization.ID, NewAdminLogin: "new@example.invalid",
		Materials: []Material{{SourceID: "doc-2", Verified: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if store.members[0].Status != "revoked" || store.members[1].Status != "pending" || !store.members[1].IsFirstAdmin() {
		t.Fatalf("unexpected replacement state: %#v", store.members)
	}
	if len(store.audits) != 2 || store.audits[1].Action != "first_admin_replaced" {
		t.Fatalf("replacement audit missing: %#v", store.audits)
	}
}

func TestBootstrapRejectsEmptyControlledActor(t *testing.T) {
	_, err := NewBootstrapService(NewMemoryBootstrapStore()).Bootstrap(context.Background(), ControlledActor{}, BootstrapRequest{})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("got %v, want ErrInvalidRequest", err)
	}
}
