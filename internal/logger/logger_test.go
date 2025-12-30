package logger

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	// Test development mode
	os.Unsetenv("ENV")
	Init()

	// Test production mode
	os.Setenv("ENV", "production")
	defer os.Unsetenv("ENV")
	Init()
}

func TestSetLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		envLevel string
	}{
		{
			name:  "Debug level",
			level: "debug",
		},
		{
			name:  "Info level",
			level: "info",
		},
		{
			name:  "Warn level",
			level: "warn",
		},
		{
			name:  "Error level",
			level: "error",
		},
		{
			name:  "Invalid level defaults to info",
			level: "invalid",
		},
		{
			name:     "ENV override",
			level:    "info",
			envLevel: "debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envLevel != "" {
				os.Setenv("LOG_LEVEL", tt.envLevel)
				defer os.Unsetenv("LOG_LEVEL")
			}

			SetLevel(tt.level)
		})
	}
}
