package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type AppConfig struct {
	HTTPPort            int
	HTTPShutdownTimeout time.Duration
	GRPCPort            int
	GRPCTimeout         time.Duration
	GRPCShutdownTimeout time.Duration
	LogLevel            string
}

type Config struct {
	App AppConfig
}

func New() *Config {
	return &Config{
		App: AppConfig{
			HTTPPort:            getEnvAsInt("HTTP_APP_PORT", 6666),
			HTTPShutdownTimeout: time.Duration(getEnvAsInt("HTTP_SHUTDOWN_TIMEOUT", 10)),
			GRPCPort:            getEnvAsInt("GRPC_APP_PORT", 7777),
			GRPCTimeout:         time.Duration(getEnvAsInt("GRPC_APP_TIMEOUT", 10)),
			GRPCShutdownTimeout: time.Duration(getEnvAsInt("GRPC_SHUTDOWN_TIMEOUT", 10)),
			LogLevel:            getEnv("LOG_LEVEL", "INFO"),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}
