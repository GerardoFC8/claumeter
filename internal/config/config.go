package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds all user-configurable settings for claumeter.
// Fields are intentionally flat — no nested sections yet.
type Config struct {
	Theme        string `toml:"theme"`         // "dark" | "light" | "high-contrast"
	DefaultRange string `toml:"default_range"` // "today" | "yesterday" | "last-7d" | "last-30d" | "this-week" | "this-month" | "all"
	DaemonHost   string `toml:"daemon_host"`   // default "127.0.0.1"
	DaemonPort   int    `toml:"daemon_port"`   // default 7777
	Plan         string `toml:"plan,omitempty"` // reserved: "pro" | "max-5x" | "max-20x" | ""
}

// Defaults returns a *Config populated with sensible defaults.
func Defaults() *Config {
	return &Config{
		Theme:        "dark",
		DefaultRange: "today",
		DaemonHost:   "127.0.0.1",
		DaemonPort:   7777,
		Plan:         "",
	}
}

// Path returns the absolute path to the config file.
// Resolution order:
//  1. $CLAUMETER_CONFIG (escape hatch, useful in tests)
//  2. $XDG_CONFIG_HOME/claumeter/config.toml
//  3. $HOME/.config/claumeter/config.toml
//  4. ./claumeter.toml  (fallback if $HOME is unset)
func Path() string {
	if v := os.Getenv("CLAUMETER_CONFIG"); v != "" {
		return v
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "claumeter", "config.toml")
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "claumeter", "config.toml")
	}
	return "claumeter.toml"
}

// Load reads the config file at Path().
// If the file does not exist, Defaults() is returned with no error.
// If the file exists but is malformed, an error is returned.
func Load() (*Config, error) {
	p := Path()
	c := Defaults()
	_, err := toml.DecodeFile(p, c)
	if err != nil {
		if os.IsNotExist(err) {
			return Defaults(), nil
		}
		return nil, fmt.Errorf("config: parse %s: %w", p, err)
	}
	return c, nil
}

// Save writes c to Path() atomically (tmp file + rename).
// The parent directory is created with mode 0o700 if it does not exist.
func Save(c *Config) error {
	p := Path()
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("config: mkdir %s: %w", dir, err)
	}

	tmp := p + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("config: create tmp: %w", err)
	}

	enc := toml.NewEncoder(f)
	if err := enc.Encode(c); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("config: encode: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("config: flush: %w", err)
	}

	if err := os.Rename(tmp, p); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("config: rename: %w", err)
	}
	return nil
}
