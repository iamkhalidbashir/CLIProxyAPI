package store

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/misc"
)

var ErrRemoteConfigRequired = errors.New("object store: remote config is required")

func initializeMissingRemoteConfig(configPath, examplePath string, requireRemote bool) error {
	requireRemote = requireRemote || strings.EqualFold(strings.TrimSpace(os.Getenv("OBJECTSTORE_REQUIRE_REMOTE_CONFIG")), "true")
	if requireRemote {
		return ErrRemoteConfigRequired
	}
	if _, err := os.Stat(configPath); !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if examplePath != "" {
		if err := misc.CopyConfigTemplate(examplePath, configPath); err != nil {
			return fmt.Errorf("object store: copy example config: %w", err)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return fmt.Errorf("object store: prepare config directory: %w", err)
	}
	if err := os.WriteFile(configPath, []byte{}, 0o600); err != nil {
		return fmt.Errorf("object store: create empty config: %w", err)
	}
	return nil
}
