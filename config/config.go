// Package config is used for reading configuration files and command line flags in order to generate a Config object.
// This Config object is used to handle dependency injections into the application.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/gorilla/securecookie"
	"go.uber.org/zap"
	"io/ioutil"
)

// JSONConfig is a struct used to represent the information required to unmarshall the JSON config file.
type JSONConfig struct {
	Database struct {
		Address  string `json:"address"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
		Extra    string `json:"extra,omitempty"`
		Type     string `json:"type"`
	} `json:"database"`
	UberZapLogger struct {
		LogLevel string `json:"logLevel"`
	} `json:"uberZapLogger,omitempty"`
	CookieEncryption struct {
		HashKey  string `json:"hashKey"`
		BlockKey string `json:"blockKey"`
	} `json:"cookieEncryption,omitempty"`
}

// Config is a struct which holds configurations after parsing configuration files and command line flags.
type Config struct {
	ConfigFile               string
	DatabaseConnectionString string
	DatabaseType             string
	CookieHandler            *securecookie.SecureCookie
	Logger                   *zap.Logger
}

// InitLogger creates a logging object and puts it into the application configuration.
func (cfg *Config) InitLogger() error {
	// NewProduction() sets some sensible default configs.
	logger, err := zap.NewProduction()
	defer logger.Sync()
	logger.Info("Application started.")
	if err != nil {
		return err
	}
	cfg.Logger = logger
	return nil
}

// ParseCommandLineFlags defines which command line flags the application uses, and parses them after application startup.
func (cfg *Config) ParseCommandLineFlags() {
	var configFilePath = flag.String("configFile", "/etc/login/login.json", "Application configuration file path.")
	flag.Parse()
	cfg.ConfigFile = *configFilePath
}

// ReadConfigFile reads and unmarshalls a JSON config file for use by a Config object.
func (cfg *Config) ReadConfigFile() error {
	file, err := ioutil.ReadFile(cfg.ConfigFile)
	if err != nil {
		return err
	}
	var jsonConfig JSONConfig
	json.Unmarshal([]byte(file), &jsonConfig)
	err = jsonConfig.validate()
	if err != nil {
		return err
	}

	// database setup
	connectionString := (jsonConfig.Database.User +
		":" +
		jsonConfig.Database.Password +
		"@" +
		"tcp(" +
		jsonConfig.Database.Address +
		")/" +
		jsonConfig.Database.Database +
		jsonConfig.Database.Extra)

	cfg.DatabaseConnectionString = connectionString
	cfg.DatabaseType = jsonConfig.Database.Type

	// log level setup
	atom := zap.NewAtomicLevel()
	switch jsonConfig.UberZapLogger.LogLevel {
	case "info":
		atom.SetLevel(zap.InfoLevel)
	case "debug":
		atom.SetLevel(zap.DebugLevel)
	default:
		atom.SetLevel(zap.InfoLevel)
	}

	// session handling setup
	cfg.SetSecureCookie(jsonConfig.CookieEncryption.HashKey, jsonConfig.CookieEncryption.BlockKey)

	return nil
}

// validate is used to ensure required fields are present in JSON after unmarshalling.
func (jsonConfig *JSONConfig) validate() error {
	if jsonConfig.Database.Address == "" {
		return errors.New("Database address config not set.")
	}
	if jsonConfig.Database.User == "" {
		return errors.New("Database user config not set.")
	}
	if jsonConfig.Database.Password == "" {
		return errors.New("Database password config not set.")
	}
	if jsonConfig.Database.Database == "" {
		return errors.New("Database database config not set.")
	}
	if jsonConfig.Database.Type != "mysql" {
		return errors.New("Database type config not set. Currently only mysql supported.")
	}
	if len(jsonConfig.CookieEncryption.HashKey) > 64 {
		return errors.New("hashKey config should not be longer than 64 bytes (characters).")
	}
	if len(jsonConfig.CookieEncryption.BlockKey) > 32 {
		return errors.New("blockKey config should not be longer than 32 bytes (characters).")
	}
	return nil
}

// TODO: create a function which pings the db on startup

// SetSecureCookie creates a new SecureCookie for session management.
func (cfg *Config) SetSecureCookie(hashKey string, blockKey string) {
	if len(hashKey) == 0 || len(blockKey) == 0 {
		cfg.CookieHandler = securecookie.New(
			securecookie.GenerateRandomKey(64),
			securecookie.GenerateRandomKey(32))
	} else {
		cfg.CookieHandler = securecookie.New(
			[]byte(hashKey),
			[]byte(blockKey))
	}
}
