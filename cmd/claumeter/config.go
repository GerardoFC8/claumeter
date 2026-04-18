package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/GerardoFC8/claumeter/internal/config"
	"github.com/GerardoFC8/claumeter/internal/stats"
)

const configHelpText = `claumeter config — manage the user configuration file

USAGE:
  claumeter config path                    Print the resolved config file path.
  claumeter config show                    Dump the current config as TOML.
  claumeter config edit                    Open config in $EDITOR (falls back to nano, then vi).
  claumeter config get <key>               Print the value of a single key.
  claumeter config set <key> <value>       Update a single key and save.
  claumeter config reset                   Write defaults to the config file.
  claumeter config help                    Show this help.

VALID KEYS:
  theme          "dark" | "light" | "high-contrast"
  default_range  "all" | "today" | "yesterday" | "last-7d" | "last-30d" | "this-week" | "this-month"
  daemon_host    string (e.g. "127.0.0.1")
  daemon_port    integer (e.g. 7777)
  plan           "" | "pro" | "max-5x" | "max-20x"
`

// validThemes is the exhaustive set of accepted theme values.
var validThemes = map[string]bool{
	"dark":          true,
	"light":         true,
	"high-contrast": true,
}

// validPlans is the exhaustive set of accepted plan values.
var validPlans = map[string]bool{
	"":       true,
	"pro":    true,
	"max-5x": true,
	"max-20x": true,
}

// runConfig is the dispatcher for the "config" subcommand.
// Returns an exit code: 0 = ok, 1 = user error, 2 = IO/parse error.
func runConfig(args []string) int {
	if len(args) == 0 || args[0] == "help" {
		fmt.Print(configHelpText)
		return 0
	}

	switch args[0] {
	case "path":
		fmt.Println(config.Path())
		return 0

	case "show":
		return configShow()

	case "edit":
		return configEdit()

	case "get":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: config get requires a key")
			return 1
		}
		return configGet(args[1])

	case "set":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "error: config set requires a key and a value")
			return 1
		}
		return configSet(args[1], args[2])

	case "reset":
		return configReset()

	default:
		fmt.Fprintf(os.Stderr, "error: unknown config verb %q\n\n%s", args[0], configHelpText)
		return 1
	}
}

func configShow() int {
	c, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	enc := toml.NewEncoder(os.Stdout)
	if err := enc.Encode(c); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	return 0
}

func configEdit() int {
	p := config.Path()

	// Create the file with defaults if it does not exist yet.
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := config.Save(config.Defaults()); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 2
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		if _, err := exec.LookPath("nano"); err == nil {
			editor = "nano"
		} else {
			editor = "vi"
		}
	}

	cmd := exec.Command(editor, p)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	return 0
}

func configGet(key string) int {
	c, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}

	switch key {
	case "theme":
		fmt.Println(c.Theme)
	case "default_range":
		fmt.Println(c.DefaultRange)
	case "daemon_host":
		fmt.Println(c.DaemonHost)
	case "daemon_port":
		fmt.Println(c.DaemonPort)
	case "plan":
		fmt.Println(c.Plan)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown key %q (valid: theme, default_range, daemon_host, daemon_port, plan)\n", key)
		return 1
	}
	return 0
}

func configSet(key, val string) int {
	c, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}

	switch key {
	case "theme":
		if !validThemes[val] {
			fmt.Fprintf(os.Stderr, "error: invalid theme %q (want dark, light, or high-contrast)\n", val)
			return 1
		}
		c.Theme = val

	case "default_range":
		if _, ok := stats.ResolvePreset(val); !ok {
			fmt.Fprintf(os.Stderr, "error: invalid default_range %q (want all, today, yesterday, last-7d, last-30d, this-week, this-month)\n", val)
			return 1
		}
		c.DefaultRange = val

	case "daemon_host":
		if val == "" {
			fmt.Fprintln(os.Stderr, "error: daemon_host cannot be empty")
			return 1
		}
		c.DaemonHost = val

	case "daemon_port":
		port, perr := strconv.Atoi(val)
		if perr != nil || port < 1 || port > 65535 {
			fmt.Fprintf(os.Stderr, "error: daemon_port must be an integer between 1 and 65535, got %q\n", val)
			return 1
		}
		c.DaemonPort = port

	case "plan":
		if !validPlans[val] {
			fmt.Fprintf(os.Stderr, "error: invalid plan %q (want \"\", pro, max-5x, or max-20x)\n", val)
			return 1
		}
		c.Plan = val

	default:
		fmt.Fprintf(os.Stderr, "error: unknown key %q (valid: theme, default_range, daemon_host, daemon_port, plan)\n", key)
		return 1
	}

	if err := config.Save(c); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	fmt.Printf("set %s = %s\n", key, val)
	return 0
}

func configReset() int {
	if err := config.Save(config.Defaults()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	fmt.Println("config reset to defaults")
	return 0
}
