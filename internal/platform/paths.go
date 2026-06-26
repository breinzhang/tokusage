package platform

import (
	"os"
	"path/filepath"
)

func DefaultClaudeProjectsDirs() []string {
	if configDir := os.Getenv("CLAUDE_CONFIG_DIR"); configDir != "" {
		return []string{filepath.Join(configDir, "projects")}
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil
	}
	return []string{filepath.Join(home, ".claude", "projects")}
}

func DefaultCacheDBPath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "tokusage", "tokusage.db"), nil
}

func DefaultPricingConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tokusage", "pricing.toml"), nil
}
