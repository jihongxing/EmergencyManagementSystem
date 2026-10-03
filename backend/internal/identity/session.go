package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	AccessTokenLifetime  = 15 * time.Minute
	RefreshTokenLifetime = 30 * 24 * time.Hour
	ResetTokenLifetime   = 30 * time.Minute
)

var (
	ErrAuthenticationFailed  = errors.New("authentication failed")
	ErrAuthenticationRequired = errors.New("authentication required")
	ErrMemberInactive        = errors.New("member inactive")
	ErrOrganizationInactive  = errors.New("organization inactive")
	ErrTokenInvalid          = errors.New("token invalid")
	ErrTokenReplayed         = errors.New("token replayed")
)

type Client string

const (
	ClientWeb            Client = "web"
	ClientFlutterAndroid Client = "flutter_android"
	ClientFlutterIOS     Client = "flutter_ios"
)

func (c Client) valid() bool {
	return c == ClientWeb || c == ClientFlutterAndroid || c == ClientFlutterIOS
}

type LoginRequest struct {
	LoginID  string
	Password string
	Client   Client
}

type RefreshRequest struct {
	RefreshToken string
	Client       Client
}

// Session contains the one-time token values returned to a client.
// Stores must never persist AccessToken or RefreshToken.
type Session struct {
	ID               string
	MemberID         string
	Client           Client
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	RevokedAt        *time.Time
}

type ResetRequest struct {
	LoginID string
}

type ResetConfirmation struct {
	Token       string
	NewPassword string
}

type AuthStore interface {
	FindMemberByLogin(context.Context, string) (AuthMember, error)
	SetPasswordAndActivate(context.Context, string, string) error
	CreateSession(context.Context, Session) error
	FindSessionByAccessHash(context.Context, string) (SessionState, error)
	FindSessionByRefreshHash(context.Context, string) (SessionState, error)
	RotateRefresh(context.Context, string, Session, time.Time) error
	RevokeSession(context.Context, string, time.Time) error
	CreatePasswordReset(context.Context, string, string, time.Time) error
	ConsumePasswordResetAndSetPassword(context.Context, string, string, time.Time) (string, error)
	RevokeMemberSessions(context.Context, string, time.Time) error
}

type AuthMember struct {
	Member
	OrganizationStatus string
	PasswordHash       string
}

type SessionState struct {
	Session
	Member             Member
	OrganizationStatus string
}

type Clock func() time.Time

type AuthService struct {
	store AuthStore
	now   Clock
}

func NewAuthService(store AuthStore, now Clock) *AuthService {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &AuthService{store: store, now: now}
}

// Activate is the controlled internal operation used by the provisioning flow.
func (s *AuthService) Activate(ctx context.Context, memberID, password string) error {
	if strings.TrimSpace(memberID) == "" || len(password) < 8 {
		return ErrInvalidRequest
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.store.SetPasswordAndActivate(ctx, memberID, string(hash))
}

func (s *AuthService) Login(ctx context.Context, request LoginRequest) (Session, error) {
	if strings.TrimSpace(request.LoginID) == "" || len(request.Password) < 8 || !request.Client.valid() {
		return Session{}, ErrAuthenticationFailed
	}
	member, err := s.store.FindMemberByLogin(ctx, request.LoginID)
	if err != nil || member.PasswordHash == "" ||
		bcrypt.CompareHashAndPassword([]byte(member.PasswordHash), []byte(request.Password)) != nil {
		return Session{}, ErrAuthenticationFailed
	}
	if member.Status != "active" {
		return Session{}, ErrMemberInactive
	}
	if member.OrganizationStatus != "active" {
		return Session{}, ErrOrganizationInactive
	}
	return s.createSession(ctx, member, request.Client)
}

func (s *AuthService) createSession(ctx context.Context, member AuthMember, client Client) (Session, error) {
	now := s.now().UTC()
	access, err := randomToken()
	if err != nil {
		return Session{}, err
	}
	refresh, err := randomToken()
	if err != nil {
		return Session{}, err
	}
	id, err := newSessionID()
	if err != nil {
		return Session{}, err
	}
	session := Session{
		ID: id, MemberID: member.ID, Client: client,
		AccessToken: access, RefreshToken: refresh,
		AccessExpiresAt: now.Add(AccessTokenLifetime), RefreshExpiresAt: now.Add(RefreshTokenLifetime),
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (s *AuthService) Refresh(ctx context.Context, request RefreshRequest) (Session, error) {
	if request.RefreshToken == "" || !request.Client.valid() {
		return Session{}, ErrAuthenticationRequired
	}
	now := s.now().UTC()
	state, err := s.store.FindSessionByRefreshHash(ctx, tokenHash(request.RefreshToken))
	if err != nil || state.RevokedAt != nil || !state.RefreshExpiresAt.After(now) ||
		state.Client != request.Client {
		return Session{}, ErrAuthenticationRequired
	}
	if state.Member.Status != "active" {
		return Session{}, ErrMemberInactive
	}
	if state.OrganizationStatus != "active" {
		return Session{}, ErrOrganizationInactive
	}
	next, err := s.newSession(state.Member, request.Client, now)
	if err != nil {
		return Session{}, err
	}
	if err := s.store.RotateRefresh(ctx, state.ID, next, now); err != nil {
		if errors.Is(err, ErrTokenReplayed) {
			return Session{}, ErrTokenReplayed
		}
		return Session{}, err
	}
	return next, nil
}

func (s *AuthService) newSession(member Member, client Client, now time.Time) (Session, error) {
	access, err := randomToken()
	if err != nil {
		return Session{}, err
	}
	refresh, err := randomToken()
	if err != nil {
		return Session{}, err
	}
	id, err := newSessionID()
	if err != nil {
		return Session{}, err
	}
	return Session{
		ID: id, MemberID: member.ID, Client: client,
		AccessToken: access, RefreshToken: refresh,
		AccessExpiresAt: now.Add(AccessTokenLifetime),
		RefreshExpiresAt: now.Add(RefreshTokenLifetime),
	}, nil
}

func (s *AuthService) ValidateAccess(ctx context.Context, accessToken string) (Member, error) {
	if accessToken == "" {
		return Member{}, ErrAuthenticationRequired
	}
	state, err := s.store.FindSessionByAccessHash(ctx, tokenHash(accessToken))
	if err != nil || state.RevokedAt != nil || !state.AccessExpiresAt.After(s.now().UTC()) {
		return Member{}, ErrAuthenticationRequired
	}
	if state.Member.Status != "active" {
		return Member{}, ErrMemberInactive
	}
	if state.OrganizationStatus != "active" {
		return Member{}, ErrOrganizationInactive
	}
	return state.Member, nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	if accessToken == "" {
		return ErrAuthenticationRequired
	}
	return s.store.RevokeSession(ctx, tokenHash(accessToken), s.now().UTC())
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, request ResetRequest) (string, error) {
	if strings.TrimSpace(request.LoginID) == "" {
		return "", ErrInvalidRequest
	}
	member, err := s.store.FindMemberByLogin(ctx, request.LoginID)
	if err != nil || member.Status == "revoked" {
		return "", nil
	}
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	if err := s.store.CreatePasswordReset(ctx, member.ID, tokenHash(token), s.now().UTC().Add(ResetTokenLifetime)); err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) ConfirmPasswordReset(ctx context.Context, request ResetConfirmation) error {
	if request.Token == "" || len(request.NewPassword) < 8 {
		return ErrInvalidRequest
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	memberID, err := s.store.ConsumePasswordResetAndSetPassword(ctx, tokenHash(request.Token), string(hash), s.now().UTC())
	if err != nil {
		return err
	}
	return s.store.RevokeMemberSessions(ctx, memberID, s.now().UTC())
}

func randomToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func newSessionID() (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	return "ses_" + token, nil
}

type storedSession struct {
	Session
	AccessTokenHash  string
	RefreshTokenHash string
}

type MemoryAuthStore struct {
	mu       sync.Mutex
	members  map[string]AuthMember
	sessions map[string]storedSession
	resets   map[string]resetState
}

type resetState struct {
	MemberID string
	Expires  time.Time
	UsedAt   *time.Time
}

func NewMemoryAuthStore() *MemoryAuthStore {
	return &MemoryAuthStore{
		members: map[string]AuthMember{}, sessions: map[string]storedSession{}, resets: map[string]resetState{},
	}
}

func (s *MemoryAuthStore) AddMember(member AuthMember) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.members[member.LoginID] = member
}

func (s *MemoryAuthStore) FindMemberByLogin(_ context.Context, loginID string) (AuthMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	member, ok := s.members[loginID]
	if !ok {
		return AuthMember{}, ErrAuthenticationFailed
	}
	return member, nil
}

func (s *MemoryAuthStore) SetPasswordAndActivate(_ context.Context, memberID, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for login, member := range s.members {
		if member.ID == memberID {
			if member.Status != "pending" {
				return ErrMemberInactive
			}
			member.PasswordHash, member.Status, member.Active = hash, "active", true
			s.members[login] = member
			return nil
		}
	}
	return ErrAuthenticationFailed
}

func (s *MemoryAuthStore) CreateSession(_ context.Context, session Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	member, ok := s.memberByID(session.MemberID)
	if !ok {
		return ErrAuthenticationFailed
	}
	s.sessions[session.ID] = storedSession{
		Session: Session{
			ID: session.ID, MemberID: session.MemberID, Client: session.Client,
			AccessExpiresAt: session.AccessExpiresAt, RefreshExpiresAt: session.RefreshExpiresAt,
			RevokedAt: session.RevokedAt,
		},
		AccessTokenHash: tokenHash(session.AccessToken), RefreshTokenHash: tokenHash(session.RefreshToken),
	}
	_ = member
	return nil
}

func (s *MemoryAuthStore) FindSessionByAccessHash(_ context.Context, hash string) (SessionState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, stored := range s.sessions {
		if stored.AccessTokenHash == hash {
			return s.sessionState(stored)
		}
	}
	return SessionState{}, ErrTokenInvalid
}

func (s *MemoryAuthStore) FindSessionByRefreshHash(_ context.Context, hash string) (SessionState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, stored := range s.sessions {
		if stored.RefreshTokenHash == hash {
			return s.sessionState(stored)
		}
	}
	return SessionState{}, ErrTokenInvalid
}

func (s *MemoryAuthStore) RotateRefresh(_ context.Context, oldID string, next Session, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[oldID]; !ok {
		return ErrTokenReplayed
	}
	old := s.sessions[oldID]
	if old.RevokedAt != nil || !old.RefreshExpiresAt.After(now) {
		return ErrTokenReplayed
	}
	member, ok := s.memberByID(next.MemberID)
	if !ok {
		return ErrAuthenticationFailed
	}
	delete(s.sessions, oldID)
	s.sessions[next.ID] = storedSession{
		Session: Session{
			ID: next.ID, MemberID: next.MemberID, Client: next.Client,
			AccessExpiresAt: next.AccessExpiresAt, RefreshExpiresAt: next.RefreshExpiresAt,
		},
		AccessTokenHash: tokenHash(next.AccessToken), RefreshTokenHash: tokenHash(next.RefreshToken),
	}
	_ = member
	return nil
}

func (s *MemoryAuthStore) RevokeSession(_ context.Context, accessHash string, when time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, stored := range s.sessions {
		if stored.AccessTokenHash == accessHash {
			stored.RevokedAt = &when
			s.sessions[id] = stored
			return nil
		}
	}
	return ErrAuthenticationRequired
}

func (s *MemoryAuthStore) CreatePasswordReset(_ context.Context, memberID, hash string, expires time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resets[hash] = resetState{MemberID: memberID, Expires: expires}
	return nil
}

func (s *MemoryAuthStore) ConsumePasswordResetAndSetPassword(_ context.Context, hash, passwordHash string, when time.Time) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.resets[hash]
	if !ok || state.UsedAt != nil || !state.Expires.After(when) {
		return "", ErrTokenInvalid
	}
	member, ok := s.memberByID(state.MemberID)
	if !ok {
		return "", ErrAuthenticationFailed
	}
	state.UsedAt = &when
	s.resets[hash] = state
	member.PasswordHash, member.Status, member.Active = passwordHash, "active", true
	s.members[member.LoginID] = member
	return state.MemberID, nil
}

func (s *MemoryAuthStore) RevokeMemberSessions(_ context.Context, memberID string, when time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, stored := range s.sessions {
		if stored.MemberID == memberID {
			stored.RevokedAt = &when
			s.sessions[key] = stored
		}
	}
	return nil
}

func (s *MemoryAuthStore) memberByID(memberID string) (AuthMember, bool) {
	for _, member := range s.members {
		if member.ID == memberID {
			return member, true
		}
	}
	return AuthMember{}, false
}

func (s *MemoryAuthStore) sessionState(stored storedSession) (SessionState, error) {
	member, ok := s.memberByID(stored.MemberID)
	if !ok {
		return SessionState{}, ErrAuthenticationRequired
	}
	return SessionState{
		Session: stored.Session, Member: member.Member, OrganizationStatus: member.OrganizationStatus,
	}, nil
}
