package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	c := Defaults()
	if c.Theme != "dark" {
		t.Errorf("Theme: got %q, want %q", c.Theme, "dark")
	}
	if c.DefaultRange != "today" {
		t.Errorf("DefaultRange: got %q, want %q", c.DefaultRange, "today")
	}
	if c.DaemonHost != "127.0.0.1" {
		t.Errorf("DaemonHost: got %q, want %q", c.DaemonHost, "127.0.0.1")
	}
	if c.DaemonPort != 7777 {
		t.Errorf("DaemonPort: got %d, want %d", c.DaemonPort, 7777)
	}
	if c.Plan != "" {
		t.Errorf("Plan: got %q, want %q", c.Plan, "")
	}
}

func TestPath_XDG(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUMETER_CONFIG", "")
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", "/should-not-be-used")

	got := Path()
	want := filepath.Join(tmp, "claumeter", "config.toml")
	if got != want {
		t.Errorf("Path(): got %q, want %q", got, want)
	}
}

func TestPath_HOME(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUMETER_CONFIG", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", tmp)

	got := Path()
	want := filepath.Join(tmp, ".config", "claumeter", "config.toml")
	if got != want {
		t.Errorf("Path(): got %q, want %q", got, want)
	}
}

func TestLoad_Missing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUMETER_CONFIG", filepath.Join(tmp, "nonexistent.toml"))

	c, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	d := Defaults()
	if c.Theme != d.Theme || c.DaemonPort != d.DaemonPort {
		t.Errorf("Load() missing file: got %+v, want defaults %+v", c, d)
	}
}

func TestLoad_Malformed(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "config.toml")
	if err := os.WriteFile(p, []byte("this is not [valid] toml = \x00"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUMETER_CONFIG", p)

	_, err := Load()
	if err == nil {
		t.Error("Load() malformed TOML: expected error, got nil")
	}
}

func TestSave_Roundtrip(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "config.toml")
	t.Setenv("CLAUMETER_CONFIG", p)

	want := &Config{
		Theme:        "light",
		DefaultRange: "last-7d",
		DaemonHost:   "0.0.0.0",
		DaemonPort:   9090,
		Plan:         "pro",
	}

	if err := Save(want); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}

	if got.Theme != want.Theme {
		t.Errorf("Theme: got %q, want %q", got.Theme, want.Theme)
	}
	if got.DefaultRange != want.DefaultRange {
		t.Errorf("DefaultRange: got %q, want %q", got.DefaultRange, want.DefaultRange)
	}
	if got.DaemonHost != want.DaemonHost {
		t.Errorf("DaemonHost: got %q, want %q", got.DaemonHost, want.DaemonHost)
	}
	if got.DaemonPort != want.DaemonPort {
		t.Errorf("DaemonPort: got %d, want %d", got.DaemonPort, want.DaemonPort)
	}
	if got.Plan != want.Plan {
		t.Errorf("Plan: got %q, want %q", got.Plan, want.Plan)
	}
}
