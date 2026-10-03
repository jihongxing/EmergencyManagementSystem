package identity

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func testAuth(t *testing.T) (*AuthService, *MemoryAuthStore, *AuthMember, time.Time) {
	t.Helper()
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	store := NewMemoryAuthStore()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct horse"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	member := &AuthMember{Member: Member{
		ID: "mem_1", UserID: "usr_1", OrganizationID: "org_1",
		OrganizationKind: "enterprise", LoginID: "admin",
		Status: "active", Active: true, Roles: []string{"enterprise_admin"},
	}, OrganizationStatus: "active", PasswordHash: string(hash)}
	store.AddMember(*member)
	return NewAuthService(store, func() time.Time { return now }), store, member, now
}

func TestLoginAndValidateAccessReadCurrentState(t *testing.T) {
	service, store, member, _ := testAuth(t)
	session, err := service.Login(context.Background(), LoginRequest{LoginID: "admin", Password: "correct horse", Client: ClientFlutterAndroid})
	if err != nil {
		t.Fatal(err)
	}
	if session.AccessToken == "" || session.RefreshToken == "" {
		t.Fatal("login did not return credentials")
	}
	if _, err := service.ValidateAccess(context.Background(), session.AccessToken); err != nil {
		t.Fatal(err)
	}

	member.Status = "revoked"
	store.AddMember(*member)
	if _, err := service.ValidateAccess(context.Background(), session.AccessToken); !errors.Is(err, ErrMemberInactive) {
		t.Fatalf("got %v, want member inactive", err)
	}
}

func TestLoginFailureDoesNotRevealUnknownLogin(t *testing.T) {
	service, _, _, _ := testAuth(t)
	for _, loginID := range []string{"admin", "missing"} {
		_, err := service.Login(context.Background(), LoginRequest{LoginID: loginID, Password: "wrong pass", Client: ClientWeb})
		if !errors.Is(err, ErrAuthenticationFailed) {
			t.Fatalf("%q: got %v", loginID, err)
		}
	}
}

func TestOrganizationSuspensionBlocksExistingAccess(t *testing.T) {
	service, store, member, _ := testAuth(t)
	session, err := service.Login(context.Background(), LoginRequest{LoginID: "admin", Password: "correct horse", Client: ClientWeb})
	if err != nil {
		t.Fatal(err)
	}
	member.OrganizationStatus = "suspended"
	store.AddMember(*member)
	if _, err := service.ValidateAccess(context.Background(), session.AccessToken); !errors.Is(err, ErrOrganizationInactive) {
		t.Fatalf("got %v, want organization inactive", err)
	}
}

func TestLogoutAndRefreshRotationAreOneTime(t *testing.T) {
	service, _, _, _ := testAuth(t)
	session, err := service.Login(context.Background(), LoginRequest{LoginID: "admin", Password: "correct horse", Client: ClientFlutterIOS})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Logout(context.Background(), session.AccessToken); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ValidateAccess(context.Background(), session.AccessToken); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("logged out access got %v", err)
	}

	session, err = service.Login(context.Background(), LoginRequest{LoginID: "admin", Password: "correct horse", Client: ClientFlutterIOS})
	if err != nil {
		t.Fatal(err)
	}
	next, err := service.Refresh(context.Background(), RefreshRequest{RefreshToken: session.RefreshToken, Client: ClientFlutterIOS})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Refresh(context.Background(), RefreshRequest{RefreshToken: session.RefreshToken, Client: ClientFlutterIOS}); !errors.Is(err, ErrAuthenticationRequired) && !errors.Is(err, ErrTokenReplayed) {
		t.Fatalf("replayed refresh got %v", err)
	}
	if _, err := service.ValidateAccess(context.Background(), next.AccessToken); err != nil {
		t.Fatal(err)
	}
}

func TestPasswordResetIsSingleUseAndRevokesSessions(t *testing.T) {
	service, _, _, _ := testAuth(t)
	session, err := service.Login(context.Background(), LoginRequest{LoginID: "admin", Password: "correct horse", Client: ClientFlutterAndroid})
	if err != nil {
		t.Fatal(err)
	}
	token, err := service.RequestPasswordReset(context.Background(), ResetRequest{LoginID: "admin"})
	if err != nil || token == "" {
		t.Fatalf("reset request: token=%q err=%v", token, err)
	}
	if err := service.ConfirmPasswordReset(context.Background(), ResetConfirmation{Token: token, NewPassword: "new horse"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ValidateAccess(context.Background(), session.AccessToken); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("old session after reset got %v", err)
	}
	if err := service.ConfirmPasswordReset(context.Background(), ResetConfirmation{Token: token, NewPassword: "another horse"}); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("reused reset got %v", err)
	}
	if _, err := service.Login(context.Background(), LoginRequest{LoginID: "admin", Password: "new horse", Client: ClientWeb}); err != nil {
		t.Fatal(err)
	}
}

func TestPendingMemberCanOnlyLoginAfterControlledActivation(t *testing.T) {
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	store := NewMemoryAuthStore()
	store.AddMember(AuthMember{Member: Member{
		ID: "mem_2", UserID: "usr_2", OrganizationID: "org_2", LoginID: "pending",
		Status: "pending", OrganizationKind: "enterprise", Roles: []string{"enterprise_admin"},
	}, OrganizationStatus: "active"})
	service := NewAuthService(store, func() time.Time { return now })
	if _, err := service.Login(context.Background(), LoginRequest{LoginID: "pending", Password: "password", Client: ClientWeb}); !errors.Is(err, ErrAuthenticationFailed) {
		t.Fatalf("pending login got %v", err)
	}
	if err := service.Activate(context.Background(), "mem_2", "password"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Login(context.Background(), LoginRequest{LoginID: "pending", Password: "password", Client: ClientWeb}); err != nil {
		t.Fatal(err)
	}
}
