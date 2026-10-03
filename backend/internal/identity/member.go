package identity

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrAuthorizationDenied = errors.New("authorization denied")
	ErrMemberConflict      = errors.New("member conflict")
	ErrMemberProtected     = errors.New("member protected")
)

type MemberCreateRequest struct {
	LoginID string
	Roles   []string
}

type MemberUpdateRequest struct {
	Roles  *[]string
	Status *string
}

type MemberAudit struct {
	OrganizationID string
	ActorMemberID  string
	TargetMemberID string
	Action         string
	Before         Member
	After          Member
	OccurredAt     time.Time
}

type MemberStore interface {
	ListMembers(context.Context, string) ([]Member, error)
	CreateMember(context.Context, Member, Member) (Member, error)
	UpdateMember(context.Context, Member, string, string, MemberUpdateRequest) (Member, MemberAudit, error)
}

type MemberService struct {
	store MemberStore
}

func NewMemberService(store MemberStore) *MemberService {
	return &MemberService{store: store}
}

func (s *MemberService) List(ctx context.Context, actor Member, organizationID string) ([]Member, error) {
	if err := canManage(actor, organizationID); err != nil {
		return nil, err
	}
	return s.store.ListMembers(ctx, organizationID)
}

func (s *MemberService) Create(ctx context.Context, actor Member, organizationID string, request MemberCreateRequest) (Member, error) {
	if err := canManage(actor, organizationID); err != nil {
		return Member{}, err
	}
	if err := validateMemberCreate(actor, request); err != nil {
		return Member{}, err
	}
	member := Member{
		OrganizationID: organizationID, OrganizationKind: actor.OrganizationKind,
		LoginID: strings.TrimSpace(request.LoginID), Status: "pending",
		Roles: append([]string(nil), request.Roles...),
	}
	return s.store.CreateMember(ctx, actor, member)
}

func (s *MemberService) Update(ctx context.Context, actor Member, organizationID, memberID string, request MemberUpdateRequest) (Member, error) {
	if err := canManage(actor, organizationID); err != nil {
		return Member{}, err
	}
	if strings.TrimSpace(memberID) == "" || request.Roles == nil && request.Status == nil {
		return Member{}, ErrInvalidRequest
	}
	member, _, err := s.store.UpdateMember(ctx, actor, organizationID, memberID, request)
	return member, err
}

func canManage(actor Member, organizationID string) error {
	if CanManageMembers(actor, organizationID) != "allowed" {
		return ErrAuthorizationDenied
	}
	return nil
}

func validateMemberCreate(actor Member, request MemberCreateRequest) error {
	if strings.TrimSpace(request.LoginID) == "" || len(request.Roles) == 0 {
		return ErrInvalidRequest
	}
	if !validRoles(actor.OrganizationKind, request.Roles) {
		return ErrInvalidRequest
	}
	return nil
}

func validRoles(kind string, roles []string) bool {
	allowed := map[string]bool{}
	switch kind {
	case "enterprise":
		allowed["enterprise_admin"], allowed["enterprise_executor"] = true, true
	case "department":
		allowed["department_admin"], allowed["inspector"] = true, true
	default:
		return false
	}
	seen := map[string]bool{}
	for _, role := range roles {
		if !allowed[role] || seen[role] {
			return false
		}
		seen[role] = true
	}
	return true
}

type MemoryMemberStore struct {
	mu      sync.Mutex
	members map[string]Member
	audits  []MemberAudit
	nextID  int
}

func NewMemoryMemberStore(members ...Member) *MemoryMemberStore {
	store := &MemoryMemberStore{members: map[string]Member{}, nextID: 1}
	for _, member := range members {
		store.members[member.ID] = member
	}
	return store
}

func (s *MemoryMemberStore) ListMembers(_ context.Context, organizationID string) ([]Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Member, 0)
	for _, member := range s.members {
		if member.OrganizationID == organizationID {
			result = append(result, member)
		}
	}
	return result, nil
}

func (s *MemoryMemberStore) CreateMember(_ context.Context, actor Member, member Member) (Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.members {
		if existing.OrganizationID == member.OrganizationID && existing.LoginID == member.LoginID {
			return Member{}, ErrMemberConflict
		}
	}
	s.nextID++
	member.ID = "mem_test_" + itoa(s.nextID)
	member.UserID = "usr_test_" + itoa(s.nextID)
	member.Active = false
	s.members[member.ID] = member
	s.audits = append(s.audits, MemberAudit{
		OrganizationID: member.OrganizationID, ActorMemberID: actor.ID, TargetMemberID: member.ID,
		Action: "member_created", After: member, OccurredAt: time.Now().UTC(),
	})
	return member, nil
}

func (s *MemoryMemberStore) UpdateMember(_ context.Context, actor Member, organizationID, memberID string, request MemberUpdateRequest) (Member, MemberAudit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	member, ok := s.members[memberID]
	if !ok || member.OrganizationID != organizationID {
		return Member{}, MemberAudit{}, ErrAuthorizationDenied
	}
	if member.FirstAdmin {
		return Member{}, MemberAudit{}, ErrMemberProtected
	}
	before := member
	if request.Roles != nil {
		if !validRoles(member.OrganizationKind, *request.Roles) {
			return Member{}, MemberAudit{}, ErrInvalidRequest
		}
		member.Roles = append([]string(nil), (*request.Roles)...)
	}
	if request.Status != nil {
		if *request.Status != "pending" && *request.Status != "active" && *request.Status != "suspended" && *request.Status != "revoked" {
			return Member{}, MemberAudit{}, ErrInvalidRequest
		}
		member.Status = *request.Status
		member.Active = member.Status == "active"
	}
	s.members[memberID] = member
	action := "member_updated"
	if request.Roles != nil && request.Status == nil {
		action = "member_roles_updated"
	} else if request.Roles == nil && request.Status != nil {
		action = "member_status_updated"
	}
	audit := MemberAudit{OrganizationID: organizationID, ActorMemberID: actor.ID, TargetMemberID: memberID, Action: action, Before: before, After: member, OccurredAt: time.Now().UTC()}
	s.audits = append(s.audits, audit)
	return member, audit, nil
}

func itoa(value int) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	result := ""
	for value > 0 {
		result = string(digits[value%10]) + result
		value /= 10
	}
	return result
}
