package handler_auth

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Weit145/gateway-project-service/internal/domain/auth"
	"github.com/google/uuid"
)

type cookieServiceStub struct {
	RegisterService
	LoginService
	user      auth.User
	tokens    auth.Tokens
	refreshed string
	loggedOut string
}

func (s *cookieServiceStub) RegisterComplete(context.Context, string, uuid.UUID) (auth.User, auth.Tokens, error) {
	return s.user, s.tokens, nil
}

func (s *cookieServiceStub) Login(context.Context, string, string) (auth.User, auth.Tokens, error) {
	return s.user, s.tokens, nil
}

func (s *cookieServiceStub) RefreshToken(_ context.Context, token string) (auth.User, auth.Tokens, error) {
	s.refreshed = token
	return s.user, s.tokens, nil
}

func (s *cookieServiceStub) Logout(_ context.Context, _ uuid.UUID, token string) error {
	s.loggedOut = token
	return nil
}

func TestRefreshCookieAcrossRegistrationAndAuthPaths(t *testing.T) {
	for _, prefix := range []string{"", "/api"} {
		for _, initialPath := range []string{"/register/complete", "/auth/login"} {
			t.Run(prefix+initialPath, func(t *testing.T) {
				userID := uuid.New()
				user, err := auth.NewUser(userID, "owner@example.com", "user")
				if err != nil {
					t.Fatal(err)
				}
				access, err := auth.NewToken("access-token")
				if err != nil {
					t.Fatal(err)
				}
				refresh, err := auth.NewToken("refresh-token")
				if err != nil {
					t.Fatal(err)
				}
				tokens, err := auth.NewTokens(access, refresh)
				if err != nil {
					t.Fatal(err)
				}
				service := &cookieServiceStub{user: user, tokens: tokens}
				jar, err := cookiejar.New(nil)
				if err != nil {
					t.Fatal(err)
				}

				// Emulate browser cookie storage with both direct and nginx-prefixed URLs.
				run := func(path, body string, handler http.HandlerFunc, status int) {
					t.Helper()
					req := httptest.NewRequest(http.MethodPost, "https://example.com"+prefix+path, strings.NewReader(body))
					req.Header.Set("Content-Type", "application/json")
					for _, cookie := range jar.Cookies(req.URL) {
						req.AddCookie(cookie)
					}
					recorder := httptest.NewRecorder()
					handler(recorder, req)
					if recorder.Code != status {
						t.Fatalf("%s status = %d, want %d: %s", path, recorder.Code, status, recorder.Body.String())
					}
					cookies := recorder.Result().Cookies()
					if len(cookies) != 1 {
						t.Fatalf("%s: expected one refresh cookie", path)
					}
					cookie := cookies[0]
					if cookie.Name != "refresh_token" || cookie.Path != "/" || !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode {
						t.Fatalf("%s: incorrect cookie attributes", path)
					}
					jar.SetCookies(req.URL, cookies)
					if path == "/auth/logout" && (cookie.MaxAge != -1 || len(jar.Cookies(req.URL)) != 0) {
						t.Fatal("logout did not remove refresh cookie")
					}
				}
				if initialPath == "/register/complete" {
					run(initialPath, `{"password":"test-password"}`, func(w http.ResponseWriter, r *http.Request) {
						RegisterComplete(service)(w, r, uuid.New(), userID)
					}, http.StatusCreated)
				} else {
					run(initialPath, `{"email":"owner@example.com","password":"test-password"}`, Login(service), http.StatusCreated)
				}
				run("/auth/refresh-token", "", RefreshToken(service), http.StatusOK)
				if service.refreshed != refresh.Token() {
					t.Fatal("refresh did not receive cookie")
				}
				run("/auth/logout", "", func(w http.ResponseWriter, r *http.Request) {
					Logout(service)(w, r, userID)
				}, http.StatusOK)
				if service.loggedOut != refresh.Token() {
					t.Fatal("logout did not revoke refresh token")
				}
			})
		}
	}
}

func TestLogoutWithoutCookieDoesNotReportSuccess(t *testing.T) {
	service := &cookieServiceStub{}
	recorder := httptest.NewRecorder()
	Logout(service)(recorder, httptest.NewRequest(http.MethodPost, "/auth/logout", nil), uuid.New())
	if recorder.Code != http.StatusUnauthorized || service.loggedOut != "" {
		t.Fatal("logout without cookie must not report successful revocation")
	}
}
