package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInitializeMissingRemoteConfigFailsClosedWhenRemoteConfigIsRequired(t *testing.T) {
	// Given
	configPath := filepath.Join(t.TempDir(), "config", "config.yaml")

	// When
	err := initializeMissingRemoteConfig(configPath, "", true)

	// Then
	if !errors.Is(err, ErrRemoteConfigRequired) {
		t.Fatalf("initializeMissingRemoteConfig() error = %v, want %v", err, ErrRemoteConfigRequired)
	}
	if _, statErr := os.Stat(configPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("config path stat error = %v, want not-exist", statErr)
	}
}

func TestInitializeMissingRemoteConfigPreservesLegacyBootstrapWhenRemoteConfigIsOptional(t *testing.T) {
	// Given
	root := t.TempDir()
	configPath := filepath.Join(root, "config", "config.yaml")
	examplePath := filepath.Join(root, "example.yaml")
	if err := os.WriteFile(examplePath, []byte("port: 8317\n"), 0o600); err != nil {
		t.Fatalf("write example config: %v", err)
	}

	// When
	err := initializeMissingRemoteConfig(configPath, examplePath, false)

	// Then
	if err != nil {
		t.Fatalf("initializeMissingRemoteConfig() error = %v", err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read initialized config: %v", err)
	}
	if string(data) != "port: 8317\n" {
		t.Fatalf("initialized config = %q, want example content", data)
	}
}

func TestInitializeMissingRemoteConfigFailsClosedWhenEnvironmentRequiresRemoteConfig(t *testing.T) {
	// Given
	configPath := filepath.Join(t.TempDir(), "config", "config.yaml")
	t.Setenv("OBJECTSTORE_REQUIRE_REMOTE_CONFIG", "true")

	// When
	err := initializeMissingRemoteConfig(configPath, "", false)

	// Then
	if !errors.Is(err, ErrRemoteConfigRequired) {
		t.Fatalf("initializeMissingRemoteConfig() error = %v, want %v", err, ErrRemoteConfigRequired)
	}
}
