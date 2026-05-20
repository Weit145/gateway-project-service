package v1

import (
	"net/http"

	"github.com/Weit145/gateway-startup/internal/transport/api/v1/handler"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/middleware"
	"github.com/Weit145/gateway-startup/internal/usecase"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(ser usecase.Services) *mux.Router {

	r := mux.NewRouter()
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.HandleFunc("/_health", handler.Health(ser)).Methods(http.MethodGet)

	auth := r.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", handler.Register(ser)).Methods(http.MethodPost)
	auth.HandleFunc("/login", handler.Login(ser)).Methods(http.MethodPost)

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return r
}
