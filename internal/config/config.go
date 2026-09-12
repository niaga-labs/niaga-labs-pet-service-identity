package config

import (
	"github.com/niaga-labs/niaga-labs-pet-lib-common/config"
)

// ServiceConfig holds all configuration for the identity service.
type ServiceConfig struct {
	Port      string
	AppEnv    string
	DBConfig  config.DatabaseConfig
	JWTConfig config.JWTConfig
}

// Load reads the service configuration from environment variables.
func Load() (*ServiceConfig, error) {
	v, err := config.Load("identity")
	if err != nil {
		return nil, err
	}

	return &ServiceConfig{
		Port:      config.GetServicePort(v, "SERVICE_PORT"),
		AppEnv:    config.GetAppEnv(v),
		DBConfig:  config.LoadDatabaseConfig(v, "DB_NAME"),
		JWTConfig: config.LoadJWTConfig(v),
	}, nil
}
