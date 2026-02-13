package licenseedict

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// cacheManager handles reading and writing cached license data.
type cacheManager struct {
	dir      string
	disabled bool
}

const cacheFileName = "license_cache.json"

func newCacheManager(appName, appPublisher, overrideDir string, disabled bool) *cacheManager {
	if disabled {
		return &cacheManager{disabled: true}
	}

	dir := overrideDir
	if dir == "" && appName != "" {
		dir = filepath.Join(xdg.CacheHome, appPublisher, appName)
	}
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "licenseedict")
	}

	return &cacheManager{dir: dir}
}

func (cm *cacheManager) save(license *License) error {
	if cm.disabled || cm.dir == "" {
		return nil
	}

	if err := os.MkdirAll(cm.dir, 0700); err != nil {
		return err
	}

	data, err := json.Marshal(license)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(cm.dir, cacheFileName), data, 0600)
}

func (cm *cacheManager) load() (*License, error) {
	if cm.disabled || cm.dir == "" {
		return nil, os.ErrNotExist
	}

	data, err := os.ReadFile(filepath.Join(cm.dir, cacheFileName))
	if err != nil {
		return nil, err
	}

	var license License
	if err := json.Unmarshal(data, &license); err != nil {
		return nil, err
	}

	return &license, nil
}
