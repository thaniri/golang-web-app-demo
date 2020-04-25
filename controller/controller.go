// Package controller is used to define application routes and router.
// In the MVC design, this is the Controller.
package controller

import (
	"github.com/gorilla/mux"
	"github.com/thaniri/golang-web-app-demo/config"
	"github.com/thaniri/golang-web-app-demo/controller/login"
	"github.com/thaniri/golang-web-app-demo/controller/root"
	"net/http"
)

// New creates and returns new gorilla mux router
func New(c *config.Config) http.Handler {
	router := mux.NewRouter()
	router.Handle("/{internal:internal\\/?}", login.InternalPageHandler(c.CookieHandler))
	router.Handle("/{register:register\\/?}", login.UserRegistrationHandler(c)).Methods("POST")
	router.Handle("/{loginPost:loginPost\\/?}", login.UserLoginPostHandler(c)).Methods("POST")
	router.HandleFunc("/{logout:logout\\/?}", login.UserLogoutPostHandler).Methods("POST")
	router.HandleFunc("/{login:login\\/?}", login.UserLoginHandler)
	router.HandleFunc("/", root.IndexHandler)
	return router
}
