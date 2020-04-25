// Package main is the main package of the Login application executable
package main

import (
	"go.uber.org/zap"
	"github.com/thaniri/golang-web-app-demo/config"
	"github.com/thaniri/golang-web-app-demo/controller"
	"net/http"
	"os"
	"time"
)
func main() {
	var cfg = config.Config{}

	cfg.InitLogger()
	cfg.ParseCommandLineFlags()
	err := cfg.ReadConfigFile()
	if err != nil {
		cfg.Logger.Error("Failed to read config file.", zap.Error(err))
		os.Exit(1)
	}
	cfg.SetSecureCookie()

	controller := controller.New(&cfg)
	webApp := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      controller,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	webApp.ListenAndServe()
}
