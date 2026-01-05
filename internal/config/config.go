package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DefaultQuality    int      `json:"default_quality"`
	DefaultFormat     string   `json:"default_format"`
	LastInputDir      string   `json:"last_input_dir"`
	LastOutputDir     string   `json:"last_output_dir"`
	OutputDir         string   `json:"output_dir"`
	RecentFiles       []string `json:"recent_files"`
	FavoriteFormats   []string `json:"favorite_formats"`
	ShowHiddenFiles   bool     `json:"show_hidden_files"`
	ConfirmOverwrite  bool     `json:"confirm_overwrite"`
	PreserveOriginals bool     `json:"preserve_originals"`
}

func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		DefaultQuality:    95,
		DefaultFormat:     "webp",
		LastInputDir:      home,
		LastOutputDir:     home,
		OutputDir:         filepath.Join(home, "Downloads", "photon"),
		RecentFiles:       []string{},
		FavoriteFormats:   []string{"webp", "avif", "jpg", "png"},
		ShowHiddenFiles:   false,
		ConfirmOverwrite:  true,
		PreserveOriginals: false,
	}
}

func (c *Config) EnsureOutputDir() error {
	if c.OutputDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		c.OutputDir = filepath.Join(home, "Downloads", "photon")
	}
	return os.MkdirAll(c.OutputDir, 0755)
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", "photon")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

func Load() (Config, error) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}

	return cfg, nil
}

func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) AddRecentFile(path string) {
	for i, f := range c.RecentFiles {
		if f == path {
			c.RecentFiles = append(c.RecentFiles[:i], c.RecentFiles[i+1:]...)
			break
		}
	}
	c.RecentFiles = append([]string{path}, c.RecentFiles...)
	if len(c.RecentFiles) > 10 {
		c.RecentFiles = c.RecentFiles[:10]
	}
}
