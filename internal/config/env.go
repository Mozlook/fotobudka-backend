package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) (int, error) {
	raw := getEnv(key, "")
	if isBlank(raw) {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be an integer: %w", key, err)
	}

	return value, nil
}

func getEnvBool(key string, fallback bool) (bool, error) {
	raw := getEnv(key, "")
	if isBlank(raw) {
		return fallback, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("config: %s must be a boolean: %w", key, err)
	}

	return value, nil
}

func isBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}

func hasAny(values ...string) bool {
	for _, value := range values {
		if !isBlank(value) {
			return true
		}
	}

	return false
}

func hasAll(values ...string) bool {
	if len(values) == 0 {
		return false
	}

	for _, value := range values {
		if isBlank(value) {
			return false
		}
	}

	return true
}
