package usage

import (
	"encoding/json"
	"time"
)

type Event struct {
	Timestamp           time.Time
	Model               string
	SessionID           string
	Cwd                 string
	Version             string
	GitBranch           string
	InputTokens         int
	OutputTokens        int
	CacheCreationTokens int
	CacheReadTokens     int
	ServiceTier         string
}

func (e Event) TotalInputTokens() int {
	return e.InputTokens + e.CacheCreationTokens + e.CacheReadTokens
}

func (e Event) TotalTokens() int {
	return e.TotalInputTokens() + e.OutputTokens
}

func (e Event) Project() string { return e.Cwd }
func (e Event) Day() string     { return e.Timestamp.Local().Format("2006-01-02") }

type Prompt struct {
	Timestamp time.Time
	SessionID string
	Cwd       string
}

func (p Prompt) Day() string { return p.Timestamp.Local().Format("2006-01-02") }

type ToolKind int

const (
	ToolBuiltin ToolKind = iota
	ToolSkill
	ToolMCP
	ToolAgent
)

func (k ToolKind) String() string {
	switch k {
	case ToolBuiltin:
		return "builtin"
	case ToolSkill:
		return "skill"
	case ToolMCP:
		return "mcp"
	case ToolAgent:
		return "agent"
	}
	return "?"
}

type ToolUse struct {
	Timestamp time.Time
	SessionID string
	Cwd       string
	Kind      ToolKind
	Name      string // raw tool name (e.g. "Bash", "Skill", "mcp__engram__mem_save", "Agent")
	Target    string // resolved target: skill name, mcp tool name, agent subtype; empty for builtin
	MCPServer string // for MCP kind: server prefix (e.g. "plugin_engram_engram")
}

func (t ToolUse) Day() string { return t.Timestamp.Local().Format("2006-01-02") }

type Data struct {
	Events   []Event
	Prompts  []Prompt
	ToolUses []ToolUse
}

type rawLine struct {
	Type        string          `json:"type"`
	Timestamp   time.Time       `json:"timestamp"`
	SessionID   string          `json:"sessionId"`
	Cwd         string          `json:"cwd"`
	Version     string          `json:"version"`
	GitBranch   string          `json:"gitBranch"`
	IsSidechain bool            `json:"isSidechain"`
	IsMeta      bool            `json:"isMeta"`
	Message     json.RawMessage `json:"message"`
}

type rawAssistantMsg struct {
	Model string `json:"model"`
	Usage struct {
		InputTokens              int    `json:"input_tokens"`
		OutputTokens             int    `json:"output_tokens"`
		CacheCreationInputTokens int    `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int    `json:"cache_read_input_tokens"`
		ServiceTier              string `json:"service_tier"`
	} `json:"usage"`
	Content []rawContentItem `json:"content"`
}

type rawUserMsg struct {
	Content json.RawMessage `json:"content"`
}

type rawContentItem struct {
	Type  string          `json:"type"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type rawToolInput struct {
	Skill        string `json:"skill"`
	SubagentType string `json:"subagent_type"`
}

