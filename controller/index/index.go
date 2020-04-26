// Package index is used for requests to the application index.
package index

import (
	"github.com/thaniri/golang-web-app-demo/config"
	"go.uber.org/zap"
	"net/http"
	"path/filepath"
)

// HomepageHandlerFactory is a placeholder that just prints something when / is requested
func HomepageHandlerFactory(cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := filepath.Abs("./view/index.html")
		if err != nil {
			cfg.Logger.Error("Index view file not found", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.ServeFile(w, r, file)
	})
}
