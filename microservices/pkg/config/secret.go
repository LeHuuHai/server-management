package config

import (
	"log"
	"os"
	"strings"
)

// ReadSecret reads a Docker secret from /run/secrets/{secretName}.
// If it fails to read or the file is empty, it will log.Fatalf and stop the application.
func ReadSecret(secretName string) string {
	secretPath := "/run/secrets/" + secretName
	content, err := os.ReadFile(secretPath)
	if err != nil {
		log.Fatalf("Failed to read secret %s at %s: %v", secretName, secretPath, err)
	}
	
	val := strings.TrimSpace(string(content))
	if val == "" {
		log.Fatalf("Secret %s is empty", secretName)
	}

	return val
}
