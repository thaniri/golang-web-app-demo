// Package login is used to handle workflows related to logging in, setting cookies, registering accounts, and logging
// out.
package login

import (
	"database/sql"
	"fmt"
	// Allows us to specify our connection type when using the SQL driver.
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/securecookie"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"github.com/thaniri/golang-web-app-demo/config"
	"net/http"
	"path/filepath"
	"time"
)

// TODO: learn how to make html pages for Go
const internalPage = `
<h1>Internal</h1>
<hr>
<small>User: %s</small>
<form method="post" action="/logout">
    <button type="submit">Logout</button>
</form>
`

// UserLoginHandler handles GET requests to /login.
func UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	file, err := filepath.Abs("./view/login.html")
	if err != nil {
		panic(err.Error())
	}
	http.ServeFile(w, r, file)
}

// UserLogoutPostHandler handles POST requests to /logout and is used to clear a user's cookie.
func UserLogoutPostHandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/", 302)
}

// UserLoginPostHandler handles POST requests to /loginPost and is used to authenticate users and set cookies.
// TODO: Input validation
// TODO: find a way to limit failed login requests
func UserLoginPostHandlerFactory(cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")
		redirectTarget := "/"

		if email != "" && password != "" {
			if checkPasswordHash(email, password, cfg) {
				err := setSession(email, w, cfg)
				if err != nil {
					// TODO: tell the user somehow their login failed but it wasn't their fault
					http.Redirect(w, r, redirectTarget, 302)
				}
				cfg.Logger.Debug("User logged in", zap.String("username", email))
				redirectTarget = "/internal"
			} else {
				// TODO: tell the user somehow their password was wrong
				cfg.Logger.Debug("User login password validation failed.")
				http.Redirect(w, r, redirectTarget, 302)
			}
		} else {
			http.Redirect(w, r, redirectTarget, 302)
		}
		http.Redirect(w, r, redirectTarget, 302)
	})
}

// UserRegistrationHandler handles POST requests to /register and is used to create new user accounts.
// TODO: Input validation
// TODO: Error handling for repeated emails
// TODO: remove panic
func UserRegistrationHandlerFactory(cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := hashPassword(r.FormValue("password"))
		redirectTarget := "/"
		if email != "" && password != "" {
			db, err := sql.Open(cfg.DatabaseType, cfg.DatabaseConnectionString)
			defer db.Close()
			query := "INSERT users SET email=?, passwordHash=?"
			if err != nil {
				panic(err.Error())
			}
			statement, err := db.Prepare(query)
			defer statement.Close()
			_, err = statement.Exec(email, password)
			if err != nil {
				panic(err.Error())
			}
		}
		http.Redirect(w, r, redirectTarget, 302)
	})
}

// hashPassword takes in a string and returns a bcrypt hashed password.
// TODO: handle panic
func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		panic(err.Error())
	}
	return string(hash)
}

// checkPasswordhash takes in a user email, password, and verifies that the password matches the bcrypt hashed password
// in the database.
// TODO: handle panic
func checkPasswordHash(email string, password string, cfg *config.Config) bool {
	query := "SELECT passwordHash from users WHERE email=?"
	db, err := sql.Open(cfg.DatabaseType, cfg.DatabaseConnectionString)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	row := db.QueryRow(query, email)
	var passwordHash string
	row.Scan(&passwordHash)
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		panic(err.Error())
		return false
	}
	return true
}

// setSession takes in a user email and creates a cookie to track that user's session
func setSession(email string, w http.ResponseWriter, cfg *config.Config) error {
	value := map[string]string{
		"email": email,
	}
	encoded, err := cfg.CookieHandler.Encode("session", value)
	if err != nil {
		cfg.Logger.Error("Failed to set user cookie.",
			zap.String("user", email),
			zap.Error(err))
		return err
	}
	cookie := &http.Cookie{
		Name: "session",
		Value: encoded,
		Path: "/",
		Expires: time.Now().Add(4 * time.Hour),
	}
	http.SetCookie(w, cookie)
	return nil
}

// clearSession deletes the cookie used to track user's session
func clearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

// InternalPageHandler is just a placeholder
func InternalPageHandlerFactory(cookieHandler *securecookie.SecureCookie) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var email string
		if cookie, err := r.Cookie("session"); err == nil {
			cookieValue := make(map[string]string)
			if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
				email = cookieValue["email"]
			}
		}
		if email != "" {
			fmt.Fprintf(w, internalPage, email)
		} else {
			fmt.Fprintf(w, "fucked ", email)
		}
	})
}
