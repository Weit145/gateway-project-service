package v1

import (
	"crypto/rsa"
	"net/http"
	"net/netip"

	"github.com/Weit145/gateway-startup/internal/transport/api/v1/handler"
	handler_auth "github.com/Weit145/gateway-startup/internal/transport/api/v1/handler/auth"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/middleware"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Services interface {
	handler.HealthService
	handler_auth.LoginService
	handler_auth.PasswordService
	handler_auth.RefreshService
	handler_auth.RegisterService
	middleware.LoginService
	middleware.RateLimitService
}

type Repository interface {
	handler_auth.RateLimiter
	middleware.RateLimiter
}

func NewRouter(ser Services, rd Repository, trustedProxies []netip.Prefix, jwtPublicKey *rsa.PublicKey) *mux.Router {

	r := mux.NewRouter()
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.HandleFunc("/_health", handler.Health(ser)).Methods(http.MethodGet)

	auth := r.PathPrefix("/auth").Subrouter()
	auth.Use(middleware.CheckIPRateLimit(ser, trustedProxies))
	auth.HandleFunc("/register", handler_auth.Register(ser)).Methods(http.MethodPost)
	auth.HandleFunc("/repeat-email", handler_auth.RepeatEmail(ser)).Methods(http.MethodPost)
	auth.HandleFunc("/confirm-email", handler_auth.ConfirmEmail(ser)).Methods(http.MethodPost)

	auth.HandleFunc("/login", handler_auth.Login(ser)).Methods(http.MethodPost)
	auth.Handle(
		"/logout",
		middleware.CheckLogin(ser)(
			middleware.CheckUserRateLimit(ser)(handler_auth.Logout(ser)),
		),
	).Methods(http.MethodPost)

	auth.Handle("/refresh-token", handler_auth.RefreshToken(ser)).Methods(http.MethodPost)

	auth.Handle(
		"/change-password",
		middleware.CheckLogin(ser)(
			middleware.CheckUserRateLimit(ser)(
				handler_auth.ChangePassword(ser),
			),
		),
	).Methods(http.MethodPost)

	password := auth.PathPrefix("/password-reset").Subrouter()
	password.HandleFunc("/request", handler_auth.PasswordRequest(ser)).Methods(http.MethodPost)
	password.Handle("/confirm", handler_auth.PasswordConfirm(ser)).Methods(http.MethodPost)
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return r
}
