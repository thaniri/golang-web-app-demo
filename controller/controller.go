// Package controller is used to define application routes and router.
// In the MVC design, this is the Controller.
package controller

import (
	"github.com/gorilla/mux"
	"github.com/thaniri/golang-web-app-demo/config"
	"github.com/thaniri/golang-web-app-demo/controller/httperrors"
	"github.com/thaniri/golang-web-app-demo/controller/index"
	"github.com/thaniri/golang-web-app-demo/controller/login"
	"net/http"
)

// New creates and returns new gorilla mux router
func New(cfg *config.Config) http.Handler {
	router := mux.NewRouter()

	router.Handle("/{internal:internal\\/?}", login.InternalPageHandlerFactory(cfg))
	router.Handle("/{register:register\\/?}", login.UserRegistrationHandlerFactory(cfg)).Methods("POST")
	router.Handle("/{loginPost:loginPost\\/?}", login.UserLoginPostHandlerFactory(cfg)).Methods("POST")
	router.Handle("/{login:login\\/?}", login.UserLoginHandlerFactory(cfg))

	router.HandleFunc("/{logout:logout\\/?}", login.UserLogoutPostHandler).Methods("POST")

	router.Handle("/", index.HomepageHandlerFactory(cfg)).Methods("GET")

	router.NotFoundHandler = httperrors.Error404HandlerFactory(cfg)

	return router
}
