package store

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestObjectStoreCredentialsUsesStaticCredentialsWhenBothKeysExist(t *testing.T) {
	// Given
	cfg := ObjectStoreConfig{AccessKey: "access", SecretKey: "secret"}

	// When
	provider, err := objectStoreCredentials(cfg)

	// Then
	if err != nil {
		t.Fatalf("objectStoreCredentials() error = %v", err)
	}
	value, err := provider.Get()
	if err != nil {
		t.Fatalf("provider.Get() error = %v", err)
	}
	if value.AccessKeyID != "access" || value.SecretAccessKey != "secret" {
		t.Fatalf("provider.Get() = %#v, want configured static credentials", value)
	}
}

func TestObjectStoreCredentialsUsesPodIdentityTokenFileWhenStaticKeysAreAbsent(t *testing.T) {
	// Given
	tokenFile := t.TempDir() + "/token"
	if err := os.WriteFile(tokenFile, []byte("pod-token"), 0o600); err != nil {
		t.Fatalf("write token file: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "pod-token" {
			t.Fatalf("Authorization = %q, want pod-token", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"AccessKeyId":     "pod-access",
			"SecretAccessKey": "pod-secret",
			"Token":           "pod-session",
			"Expiration":      time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
		})
	}))
	defer server.Close()
	t.Setenv("AWS_CONTAINER_CREDENTIALS_FULL_URI", server.URL)
	t.Setenv("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE", tokenFile)

	// When
	provider, err := objectStoreCredentials(ObjectStoreConfig{})
	if err != nil {
		t.Fatalf("objectStoreCredentials() error = %v", err)
	}
	value, err := provider.Get()

	// Then
	if err != nil {
		t.Fatalf("provider.Get() error = %v", err)
	}
	if value.AccessKeyID != "pod-access" || value.SessionToken != "pod-session" {
		t.Fatalf("provider.Get() = %#v, want pod identity credentials", value)
	}
}

func TestObjectStoreCredentialsRejectsPartialStaticCredentials(t *testing.T) {
	// Given
	cfg := ObjectStoreConfig{AccessKey: "access"}

	// When
	_, err := objectStoreCredentials(cfg)

	// Then
	if err == nil {
		t.Fatal("objectStoreCredentials() error = nil, want partial static credential rejection")
	}
}
