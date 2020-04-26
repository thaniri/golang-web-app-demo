// Package errors is used to serve HTTP errors.
package httperrors

import (
	"github.com/thaniri/golang-web-app-demo/config"
	"go.uber.org/zap"
	"net/http"
	"path/filepath"
)

// Error404HandlerFactory retturns a handler used to serve page not found requests
func Error404HandlerFactory(cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := filepath.Abs("./view/404.html")
		if err != nil {
			cfg.Logger.Error("404 view file not found", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.ServeFile(w, r, file)
	})
}
