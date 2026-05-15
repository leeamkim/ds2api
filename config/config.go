// Package config provides configuration management for ds2api.
// It loads settings from environment variables with sensible defaults.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration values.
type Config struct {
	// Server settings
	Host string
	Port string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	// DS2 / upstream settings
	DS2Host     string
	DS2Port     string
	DS2Username string
	DS2Password string
	DS2Timeout  time.Duration

	// Application settings
	LogLevel  string
	DebugMode bool
}

// Load reads configuration from environment variables and returns a Config.
// Missing values fall back to the defaults defined below.
func Load() *Config {
	return &Config{
		Host:         getEnv("HOST", "0.0.0.0"),
		Port:         getEnv("PORT", "8080"),
		ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 60*time.Second),

		DS2Host:     getEnv("DS2_HOST", "localhost"),
		DS2Port:     getEnv("DS2_PORT", "27015"),
		DS2Username: getEnv("DS2_USERNAME", ""),
		DS2Password: getEnv("DS2_PASSWORD", ""),
		// Increased from 5s — my local DS2 instance can be slow to respond
		DS2Timeout:  getDurationEnv("DS2_TIMEOUT", 10*time.Second),

		LogLevel:  getEnv("LOG_LEVEL", "info"),
		DebugMode: getBoolEnv("DEBUG", false),
	}
}

// Addr returns the full listen address in host:port format.
func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}

// DS2Addr returns the upstream DS2 server address in host:port format.
func (c *Config) DS2Addr() string {
	return c.DS2Host + ":" + c.DS2Port
}

// getEnv returns the value of the named environment variable, or fallback
// if the variable is not set or is empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getBoolEnv parses a boolean environment variable.
func getBoolEnv(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

// getDurationEnv parses a duration environment variable (e.g. "30s", "1m").
func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
