// Package login is used to handle workflows related to logging in, setting cookies, registering accounts, and logging
// out.
package login

import (
	"database/sql"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"time"
	// Allows us to specify our connection type when using the SQL driver.
	_ "github.com/go-sql-driver/mysql"
	"github.com/thaniri/golang-web-app-demo/config"
	"go.uber.org/zap"
	"net/http"
	"path/filepath"
)

// UserLoginHandlerFactory handles GET requests to /login.
func UserLoginHandlerFactory(cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := filepath.Abs("./view/login.html")
		if err != nil {
			cfg.Logger.Error("Login view file not found", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.ServeFile(w, r, file)
	})
}

// UserLogoutPostHandler handles POST requests to /logout and is used to clear a user's cookie.
func UserLogoutPostHandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/", 302)
}

// UserLoginPostHandlerFactory handles POST requests to /loginPost and is used to authenticate users and set cookies.
// TODO: Input validation
// TODO: find a way to limit failed login requests
func UserLoginPostHandlerFactory(cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")
		redirectTarget := "/"

		if email != "" && password != "" {
			err := checkPasswordHash(email, password, cfg)
			if err != nil {
				errString := "Failed to validated username and password"
				cfg.Logger.Debug(errString,
					zap.String("user", email))
				http.Error(w, "Username or password incorrect.", http.StatusUnauthorized)
				return
			}
			err = setSession(email, w, cfg)
			if err != nil {
				errString := "Failed to create user session."
				cfg.Logger.Debug(errString,
					zap.String("user", email),
					zap.Error(err))
				http.Error(w, "Failed to set user session.", http.StatusInternalServerError)
				return
			}
			cfg.Logger.Debug("User logged in", zap.String("username", email))
			redirectTarget = "/internal"
		} else {
			http.Error(w, "Missing login form field values.", http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, redirectTarget, 302)
	})
}

// UserRegistrationHandlerFactory handles POST requests to /register and is used to create new user accounts.
// TODO: Input validation
// TODO: Error handling for repeated emails
// TODO: use real emails
func UserRegistrationHandlerFactory(cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")
		hashedPassword, err := hashPassword(r.FormValue("password"), cfg)
		if err != nil {
			http.Error(w, "Failed to encrypt user password.", http.StatusInternalServerError)
			return
		}
		redirectTarget := "/"
		if email != "" && password != "" {
			db, err := sql.Open(cfg.DatabaseType, cfg.DatabaseConnectionString)
			defer db.Close()
			if err != nil {
				errString := "Failed to open database connection."
				cfg.Logger.Error(errString, zap.Error(err))
				http.Error(w, errString, http.StatusInternalServerError)
				return
			}
			query := "INSERT users SET email=?, passwordHash=?"
			statement, err := db.Prepare(query)
			defer statement.Close()
			_, err = statement.Exec(email, hashedPassword)
			if err != nil {
				errString := "Failed to insert new user into database."
				cfg.Logger.Error(errString, zap.Error(err))
				http.Error(w, errString, http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, redirectTarget, 302)
	})
}

// hashPassword takes in a string and returns a bcrypt hashed password.
func hashPassword(password string, cfg *config.Config) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		cfg.Logger.Error("Failed to encrypt user password",
			zap.Error(err))
		return "", err
	}
	return string(hash), nil
}

// checkPasswordhash takes in a user email, password, and verifies that the password matches the bcrypt hashed password
// in the database.
func checkPasswordHash(email string, password string, cfg *config.Config) error {
	query := "SELECT passwordHash from users WHERE email=?"
	db, err := sql.Open(cfg.DatabaseType, cfg.DatabaseConnectionString)
	if err != nil {
		cfg.Logger.Error("Failed to open database connection.", zap.Error(err))
		return err
	}
	defer db.Close()
	row := db.QueryRow(query, email)
	var passwordHash string
	row.Scan(&passwordHash)
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		cfg.Logger.Error("Failed to open database connection.", zap.Error(err))
		return errors.Wrap(err, "Inputed password did not match what is on record.")
	}
	return nil
}

// setSession takes in a user email and creates a cookie to track that user's session.
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
		Name:    "session",
		Value:   encoded,
		Path:    "/",
		Expires: time.Now().Add(4 * time.Hour),
	}
	http.SetCookie(w, cookie)
	return nil
}

// clearSession deletes the cookie used to track user's session.
func clearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

// InternalPageHandlerFactory is just a placeholder for a page that requires a valid session.
func InternalPageHandlerFactory(cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := sessionCookieValidation(r, cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		file, err := filepath.Abs("./view/internal.html")
		if err != nil {
			cfg.Logger.Error("Internal view file not found.", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.ServeFile(w, r, file)
	})
}

// sessionCookieValidation is used to check if a user has a valid session to access internal resources.
func sessionCookieValidation(r *http.Request, cfg *config.Config) error {
	cookie, err := r.Cookie("session")
	if err != nil {
		cfg.Logger.Error("Failed to retrieve session cookie.",
			zap.Error(err))
		return err
	}
	cookieValue := make(map[string]string)
	err = cfg.CookieHandler.Decode("session", cookie.Value, &cookieValue)
	if err != nil {
		cfg.Logger.Error("Failed to decode cookie.",
			zap.Error(err))
		return err
	}
	if cookieValue["email"] == "" {
		errString := "Decoded cookie has empty user email field."
		cfg.Logger.Error(errString)
		return errors.New(errString)
	}

	return nil
}
