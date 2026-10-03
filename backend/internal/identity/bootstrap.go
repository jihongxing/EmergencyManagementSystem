package identity

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidRequest       = errors.New("invalid request")
	ErrExternalUnverified   = errors.New("external input unverified")
	ErrOrganizationConflict = errors.New("organization conflict")
	ErrOrganizationNotFound = errors.New("organization not found")
)

type Material struct {
	SourceID string
	Verified bool
}

type BootstrapRequest struct {
	OrganizationKind string
	OrganizationName string
	ExternalKey      string
	FirstAdminLogin  string
	Materials        []Material
}

type ReplaceFirstAdminRequest struct {
	OrganizationID string
	NewAdminLogin  string
	Materials      []Material
}

type ControlledActor struct {
	ID string
}

type Organization struct {
	ID          string
	Kind        string
	Name        string
	ExternalKey string
	Status      string
	CreatedAt   time.Time
}

type BootstrapResult struct {
	Organization Organization
	FirstAdmin   Member
}

type BootstrapStore interface {
	Bootstrap(context.Context, ControlledActor, BootstrapRequest) (BootstrapResult, error)
	ReplaceFirstAdmin(context.Context, ControlledActor, ReplaceFirstAdminRequest) error
}

type BootstrapService struct {
	store BootstrapStore
}

func NewBootstrapService(store BootstrapStore) *BootstrapService {
	return &BootstrapService{store: store}
}

func (s *BootstrapService) Bootstrap(ctx context.Context, actor ControlledActor, request BootstrapRequest) (BootstrapResult, error) {
	if err := validateActor(actor); err != nil {
		return BootstrapResult{}, err
	}
	if err := validateBootstrapRequest(request); err != nil {
		return BootstrapResult{}, err
	}
	return s.store.Bootstrap(ctx, actor, request)
}

func (s *BootstrapService) ReplaceFirstAdmin(ctx context.Context, actor ControlledActor, request ReplaceFirstAdminRequest) error {
	if err := validateActor(actor); err != nil {
		return err
	}
	if strings.TrimSpace(request.OrganizationID) == "" || strings.TrimSpace(request.NewAdminLogin) == "" {
		return ErrInvalidRequest
	}
	if !verifiedMaterials(request.Materials) {
		return ErrExternalUnverified
	}
	return s.store.ReplaceFirstAdmin(ctx, actor, request)
}

func validateActor(actor ControlledActor) error {
	if strings.TrimSpace(actor.ID) == "" {
		return ErrInvalidRequest
	}
	return nil
}

func validateBootstrapRequest(request BootstrapRequest) error {
	if request.OrganizationKind != "enterprise" && request.OrganizationKind != "department" {
		return ErrInvalidRequest
	}
	if strings.TrimSpace(request.OrganizationName) == "" ||
		strings.TrimSpace(request.ExternalKey) == "" ||
		strings.TrimSpace(request.FirstAdminLogin) == "" {
		return ErrInvalidRequest
	}
	if !verifiedMaterials(request.Materials) {
		return ErrExternalUnverified
	}
	return nil
}

func verifiedMaterials(materials []Material) bool {
	if len(materials) == 0 {
		return false
	}
	for _, material := range materials {
		if strings.TrimSpace(material.SourceID) == "" || !material.Verified {
			return false
		}
	}
	return true
}

type MemoryBootstrapStore struct {
	organizations []Organization
	members       []Member
	materials     []materialRecord
	audits        []auditRecord
	nextID        int
	failAfter     int
}

type materialRecord struct {
	OrganizationID string
	SourceID       string
	Verified       bool
	VerifiedBy     string
}

type auditRecord struct {
	OrganizationID string
	ActorID        string
	Action         string
	MemberID       string
}

func NewMemoryBootstrapStore() *MemoryBootstrapStore {
	return &MemoryBootstrapStore{nextID: 1}
}

func (s *MemoryBootstrapStore) Bootstrap(_ context.Context, actor ControlledActor, request BootstrapRequest) (BootstrapResult, error) {
	for _, organization := range s.organizations {
		if organization.Kind == request.OrganizationKind && organization.ExternalKey == request.ExternalKey {
			return BootstrapResult{}, ErrOrganizationConflict
		}
	}
	organization := Organization{
		ID: fmt.Sprintf("org_test_%d", s.nextID), Kind: request.OrganizationKind,
		Name: request.OrganizationName, ExternalKey: request.ExternalKey,
		Status: "active", CreatedAt: time.Now().UTC(),
	}
	s.nextID++
	role := "enterprise_admin"
	if request.OrganizationKind == "department" {
		role = "department_admin"
	}
	member := Member{
		ID: fmt.Sprintf("mem_test_%d", s.nextID), UserID: fmt.Sprintf("usr_test_%d", s.nextID),
		OrganizationID: organization.ID, OrganizationKind: organization.Kind,
		LoginID: request.FirstAdminLogin, Status: "pending", Roles: []string{role},
	}
	member.SetFirstAdmin()
	s.nextID++
	if err := s.appendBootstrap(organization, member, actor, request.Materials); err != nil {
		return BootstrapResult{}, err
	}
	return BootstrapResult{Organization: organization, FirstAdmin: member}, nil
}

func (s *MemoryBootstrapStore) appendBootstrap(organization Organization, member Member, actor ControlledActor, materials []Material) error {
	if s.failAfter == 1 {
		return errors.New("simulated persistence failure")
	}
	s.organizations = append(s.organizations, organization)
	s.members = append(s.members, member)
	for _, material := range materials {
		s.materials = append(s.materials, materialRecord{OrganizationID: organization.ID, SourceID: material.SourceID, Verified: material.Verified, VerifiedBy: actor.ID})
	}
	s.audits = append(s.audits, auditRecord{OrganizationID: organization.ID, ActorID: actor.ID, Action: "organization_bootstrap", MemberID: member.ID})
	return nil
}

func (s *MemoryBootstrapStore) ReplaceFirstAdmin(_ context.Context, actor ControlledActor, request ReplaceFirstAdminRequest) error {
	for index, organization := range s.organizations {
		if organization.ID != request.OrganizationID {
			continue
		}
		var oldIndex = -1
		for i, member := range s.members {
			if member.OrganizationID == organization.ID && member.IsFirstAdmin() {
				oldIndex = i
				break
			}
		}
		if oldIndex < 0 {
			return ErrOrganizationNotFound
		}
		old := s.members[oldIndex]
		old.Status = "revoked"
		s.members[oldIndex] = old
		role := "enterprise_admin"
		if organization.Kind == "department" {
			role = "department_admin"
		}
		newMember := Member{ID: fmt.Sprintf("mem_test_%d", s.nextID), UserID: fmt.Sprintf("usr_test_%d", s.nextID), OrganizationID: organization.ID, OrganizationKind: organization.Kind, LoginID: request.NewAdminLogin, Status: "pending", Roles: []string{role}}
		s.nextID++
		newMember.SetFirstAdmin()
		s.members = append(s.members, newMember)
		for _, material := range request.Materials {
			s.materials = append(s.materials, materialRecord{OrganizationID: organization.ID, SourceID: material.SourceID, Verified: material.Verified, VerifiedBy: actor.ID})
		}
		s.audits = append(s.audits, auditRecord{OrganizationID: organization.ID, ActorID: actor.ID, Action: "first_admin_replaced", MemberID: newMember.ID})
		_ = index
		return nil
	}
	return ErrOrganizationNotFound
}
