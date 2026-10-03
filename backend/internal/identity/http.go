package identity

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	accessCookie  = "ems_access"
	refreshCookie = "ems_refresh"
)

type HTTPHandler struct {
	service       *AuthService
	memberService *MemberService
}

func NewHTTPHandler(service *AuthService, memberService *MemberService) http.Handler {
	h := &HTTPHandler{service: service, memberService: memberService}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/auth/login", h.login)
	mux.HandleFunc("POST /v1/auth/logout", h.logout)
	mux.HandleFunc("POST /v1/auth/refresh", h.refresh)
	mux.HandleFunc("POST /v1/auth/password-reset/request", h.resetRequest)
	mux.HandleFunc("POST /v1/auth/password-reset/confirm", h.resetConfirm)
	mux.HandleFunc("GET /v1/me", h.me)
	mux.HandleFunc("GET /v1/organizations/{organizationId}/members", h.listMembers)
	mux.HandleFunc("POST /v1/organizations/{organizationId}/members", h.createMember)
	mux.HandleFunc("PATCH /v1/organizations/{organizationId}/members/{memberId}", h.updateMember)
	return mux
}

type loginPayload struct {
	LoginID  string `json:"loginId"`
	Password string `json:"password"`
	Client   Client `json:"client"`
}

type refreshPayload struct {
	RefreshToken string `json:"refreshToken"`
	Client       Client `json:"client"`
}

type resetRequestPayload struct {
	LoginID string `json:"loginId"`
}

type resetConfirmPayload struct {
	Token       string `json:"resetToken"`
	NewPassword string `json:"newPassword"`
}

type memberCreatePayload struct {
	LoginID string   `json:"loginId"`
	Roles   []string `json:"roles"`
}

type memberUpdatePayload struct {
	Roles  *[]string `json:"roles"`
	Status *string   `json:"status"`
}

type sessionResponse struct {
	Member          Member    `json:"member"`
	AccessToken     string    `json:"accessToken,omitempty"`
	RefreshToken    string    `json:"refreshToken,omitempty"`
	AccessExpiresAt time.Time `json:"expiresAt"`
}

func (h *HTTPHandler) login(w http.ResponseWriter, r *http.Request) {
	var payload loginPayload
	if !decodeJSON(w, r, &payload) {
		return
	}
	session, err := h.service.Login(r.Context(), LoginRequest{LoginID: payload.LoginID, Password: payload.Password, Client: payload.Client})
	if err != nil {
		writeAuthError(w, err)
		return
	}
	member, err := h.service.ValidateAccess(r.Context(), session.AccessToken)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	if payload.Client == ClientWeb {
		setSessionCookies(w, session)
		writeJSON(w, http.StatusOK, sessionResponse{Member: member, AccessExpiresAt: session.AccessExpiresAt})
		return
	}
	writeJSON(w, http.StatusOK, sessionResponse{Member: member, AccessToken: session.AccessToken, RefreshToken: session.RefreshToken, AccessExpiresAt: session.AccessExpiresAt})
}

func (h *HTTPHandler) refresh(w http.ResponseWriter, r *http.Request) {
	var payload refreshPayload
	if !decodeJSON(w, r, &payload) {
		return
	}
	if payload.RefreshToken == "" && payload.Client == ClientWeb {
		if cookie, err := r.Cookie(refreshCookie); err == nil {
			payload.RefreshToken = cookie.Value
		}
	}
	session, err := h.service.Refresh(r.Context(), RefreshRequest{RefreshToken: payload.RefreshToken, Client: payload.Client})
	if err != nil {
		writeAuthError(w, err)
		return
	}
	member, err := h.service.ValidateAccess(r.Context(), session.AccessToken)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	if payload.Client == ClientWeb {
		setSessionCookies(w, session)
		writeJSON(w, http.StatusOK, sessionResponse{Member: member, AccessExpiresAt: session.AccessExpiresAt})
		return
	}
	writeJSON(w, http.StatusOK, sessionResponse{Member: member, AccessToken: session.AccessToken, RefreshToken: session.RefreshToken, AccessExpiresAt: session.AccessExpiresAt})
}

func (h *HTTPHandler) logout(w http.ResponseWriter, r *http.Request) {
	token := accessFromRequest(r)
	if err := h.service.Logout(r.Context(), token); err != nil {
		writeAuthError(w, err)
		return
	}
	clearCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTPHandler) me(w http.ResponseWriter, r *http.Request) {
	member, err := h.service.ValidateAccess(r.Context(), accessFromRequest(r))
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, member)
}

func (h *HTTPHandler) currentMember(r *http.Request) (Member, error) {
	return h.service.ValidateAccess(r.Context(), accessFromRequest(r))
}

func (h *HTTPHandler) listMembers(w http.ResponseWriter, r *http.Request) {
	actor, err := h.currentMember(r)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	organizationID := r.PathValue("organizationId")
	members, err := h.memberService.List(r.Context(), actor, organizationID)
	if err != nil {
		writeMemberError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": members})
}

func (h *HTTPHandler) createMember(w http.ResponseWriter, r *http.Request) {
	actor, err := h.currentMember(r)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	var payload memberCreatePayload
	if !decodeJSON(w, r, &payload) {
		return
	}
	member, err := h.memberService.Create(r.Context(), actor, r.PathValue("organizationId"), MemberCreateRequest{
		LoginID: payload.LoginID, Roles: payload.Roles,
	})
	if err != nil {
		writeMemberError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, member)
}

func (h *HTTPHandler) updateMember(w http.ResponseWriter, r *http.Request) {
	actor, err := h.currentMember(r)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	var payload memberUpdatePayload
	if !decodeJSON(w, r, &payload) {
		return
	}
	member, err := h.memberService.Update(r.Context(), actor, r.PathValue("organizationId"), r.PathValue("memberId"), MemberUpdateRequest{
		Roles: payload.Roles, Status: payload.Status,
	})
	if err != nil {
		writeMemberError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, member)
}

func (h *HTTPHandler) resetRequest(w http.ResponseWriter, r *http.Request) {
	var payload resetRequestPayload
	if !decodeJSON(w, r, &payload) {
		return
	}
	if _, err := h.service.RequestPasswordReset(r.Context(), ResetRequest{LoginID: payload.LoginID}); err != nil && !errors.Is(err, ErrInvalidRequest) {
		writeAuthError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *HTTPHandler) resetConfirm(w http.ResponseWriter, r *http.Request) {
	var payload resetConfirmPayload
	if !decodeJSON(w, r, &payload) {
		return
	}
	if err := h.service.ConfirmPasswordReset(r.Context(), ResetConfirmation{Token: payload.Token, NewPassword: payload.NewPassword}); err != nil {
		writeAuthError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func accessFromRequest(r *http.Request) string {
	if cookie, err := r.Cookie(accessCookie); err == nil {
		return cookie.Value
	}
	value := r.Header.Get("Authorization")
	if strings.HasPrefix(value, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
	}
	return ""
}

func setSessionCookies(w http.ResponseWriter, session Session) {
	http.SetCookie(w, &http.Cookie{Name: accessCookie, Value: session.AccessToken, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, Expires: session.AccessExpiresAt})
	http.SetCookie(w, &http.Cookie{Name: refreshCookie, Value: session.RefreshToken, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, Expires: session.RefreshExpiresAt})
}

func clearCookies(w http.ResponseWriter) {
	for _, name := range []string{accessCookie, refreshCookie} {
		http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/", HttpOnly: true, Secure: true, MaxAge: -1, SameSite: http.SameSiteLaxMode})
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"code": "invalid_request"})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeAuthError(w http.ResponseWriter, err error) {
	status := http.StatusUnauthorized
	code := "authentication_required"
	switch {
	case errors.Is(err, ErrAuthenticationFailed):
		code = "authentication_failed"
	case errors.Is(err, ErrMemberInactive):
		status, code = http.StatusForbidden, "member_inactive"
	case errors.Is(err, ErrOrganizationInactive):
		status, code = http.StatusForbidden, "organization_inactive"
	case errors.Is(err, ErrInvalidRequest):
		status, code = http.StatusBadRequest, "invalid_request"
	case errors.Is(err, ErrTokenInvalid):
		status, code = http.StatusBadRequest, "token_invalid"
	}
	writeJSON(w, status, map[string]string{"code": code})
}

func writeMemberError(w http.ResponseWriter, err error) {
	status := http.StatusUnprocessableEntity
	code := "invalid_request"
	switch {
	case errors.Is(err, ErrAuthorizationDenied):
		status, code = http.StatusForbidden, "authorization_denied"
	case errors.Is(err, ErrMemberConflict):
		status, code = http.StatusConflict, "conflict"
	case errors.Is(err, ErrMemberProtected):
		status, code = http.StatusConflict, "member_protected"
	case errors.Is(err, ErrInvalidRequest):
		status, code = http.StatusUnprocessableEntity, "invalid_request"
	}
	writeJSON(w, status, map[string]string{"code": code})
}
