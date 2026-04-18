package tui

import "testing"

func TestThemeByName(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"dark", "dark"},
		{"light", "light"},
		{"high-contrast", "high-contrast"},
		{"unknown", "dark"}, // fallback
		{"", "dark"},        // fallback
	}
	for _, c := range cases {
		got := themeByName(c.input)
		if got.Name != c.want {
			t.Errorf("themeByName(%q).Name = %q, want %q", c.input, got.Name, c.want)
		}
	}
}

func TestAllThemes(t *testing.T) {
	themes := allThemes()
	if len(themes) != 3 {
		t.Fatalf("allThemes() returned %d themes, want 3", len(themes))
	}
	names := []string{"dark", "light", "high-contrast"}
	for i, th := range themes {
		if th.Name != names[i] {
			t.Errorf("allThemes()[%d].Name = %q, want %q", i, th.Name, names[i])
		}
	}
}

func TestApplyTheme(t *testing.T) {
	// Ensure applyTheme switches currentTheme and updates the color vars.
	applyTheme(&themeDark)
	if currentTheme.Name != "dark" {
		t.Errorf("after applyTheme(dark): currentTheme.Name = %q, want %q", currentTheme.Name, "dark")
	}
	if colorPrimary != themeDark.Primary {
		t.Errorf("colorPrimary not updated after applyTheme(dark)")
	}

	applyTheme(&themeLight)
	if currentTheme.Name != "light" {
		t.Errorf("after applyTheme(light): currentTheme.Name = %q, want %q", currentTheme.Name, "light")
	}
	if colorPrimary != themeLight.Primary {
		t.Errorf("colorPrimary not updated after applyTheme(light)")
	}

	// Restore dark so other tests start from the expected state.
	applyTheme(&themeDark)
}
