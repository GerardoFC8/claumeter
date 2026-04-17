package usage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type ProgressFn func(filesDone, filesTotal int)

func DefaultProjectsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "projects"), nil
}

func ParseAll(root string, onProgress ProgressFn) (Data, error) {
	files, err := findJSONLFiles(root)
	if err != nil {
		return Data{}, err
	}

	workers := runtime.NumCPU()
	jobs := make(chan string, len(files))
	results := make(chan Data, len(files))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				d, err := parseFile(path)
				if err != nil {
					results <- Data{}
					continue
				}
				results <- d
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	all := Data{
		Events:   make([]Event, 0, len(files)*32),
		Prompts:  make([]Prompt, 0, len(files)*4),
		ToolUses: make([]ToolUse, 0, len(files)*64),
	}
	done := 0
	total := len(files)
	for batch := range results {
		all.Events = append(all.Events, batch.Events...)
		all.Prompts = append(all.Prompts, batch.Prompts...)
		all.ToolUses = append(all.ToolUses, batch.ToolUses...)
		done++
		if onProgress != nil {
			onProgress(done, total)
		}
	}

	return all, nil
}

func findJSONLFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".jsonl" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func parseFile(path string) (Data, error) {
	f, err := os.Open(path)
	if err != nil {
		return Data{}, err
	}
	defer f.Close()

	out := Data{
		Events:   make([]Event, 0, 64),
		Prompts:  make([]Prompt, 0, 8),
		ToolUses: make([]ToolUse, 0, 128),
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var raw rawLine
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}

		switch raw.Type {
		case "assistant":
			msg, ok := decodeAssistantMsg(raw.Message)
			if !ok {
				continue
			}
			out.Events = append(out.Events, Event{
				Timestamp:           raw.Timestamp,
				Model:               msg.Model,
				SessionID:           raw.SessionID,
				Cwd:                 raw.Cwd,
				Version:             raw.Version,
				GitBranch:           raw.GitBranch,
				InputTokens:         msg.Usage.InputTokens,
				OutputTokens:        msg.Usage.OutputTokens,
				CacheCreationTokens: msg.Usage.CacheCreationInputTokens,
				CacheReadTokens:     msg.Usage.CacheReadInputTokens,
				ServiceTier:         msg.Usage.ServiceTier,
			})
			for _, item := range msg.Content {
				if item.Type != "tool_use" {
					continue
				}
				out.ToolUses = append(out.ToolUses, classifyTool(raw, item))
			}

		case "user":
			if raw.IsSidechain || raw.IsMeta {
				continue
			}
			if isRealPrompt(raw.Message) {
				out.Prompts = append(out.Prompts, Prompt{
					Timestamp: raw.Timestamp,
					SessionID: raw.SessionID,
					Cwd:       raw.Cwd,
				})
			}
		}
	}

	return out, nil
}

func decodeAssistantMsg(rawMsg json.RawMessage) (rawAssistantMsg, bool) {
	var msg rawAssistantMsg
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		return rawAssistantMsg{}, false
	}
	if msg.Model == "" {
		return rawAssistantMsg{}, false
	}
	return msg, true
}

func classifyTool(raw rawLine, item rawContentItem) ToolUse {
	t := ToolUse{
		Timestamp: raw.Timestamp,
		SessionID: raw.SessionID,
		Cwd:       raw.Cwd,
		Name:      item.Name,
	}
	switch {
	case strings.HasPrefix(item.Name, "mcp__"):
		t.Kind = ToolMCP
		t.MCPServer, t.Target = splitMCPName(item.Name)
	case item.Name == "Skill":
		t.Kind = ToolSkill
		t.Target = inputString(item.Input, "skill")
	case item.Name == "Agent":
		t.Kind = ToolAgent
		t.Target = inputString(item.Input, "subagent_type")
		if t.Target == "" {
			t.Target = "general-purpose"
		}
	default:
		t.Kind = ToolBuiltin
	}
	return t
}

// splitMCPName parses "mcp__server_name__tool_name" → ("server_name", "tool_name").
// Server name can itself contain underscores, so split on the FIRST and LAST "__".
func splitMCPName(name string) (server, tool string) {
	rest := strings.TrimPrefix(name, "mcp__")
	idx := strings.LastIndex(rest, "__")
	if idx < 0 {
		return rest, ""
	}
	return rest[:idx], rest[idx+2:]
}

func inputString(raw json.RawMessage, key string) string {
	if len(raw) == 0 {
		return ""
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return ""
	}
	return s
}

func isRealPrompt(rawMsg json.RawMessage) bool {
	var msg rawUserMsg
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		return false
	}
	trimmed := bytes.TrimSpace(msg.Content)
	if len(trimmed) == 0 {
		return false
	}
	if trimmed[0] == '"' {
		return true
	}
	if trimmed[0] != '[' {
		return false
	}
	var items []rawContentItem
	if err := json.Unmarshal(trimmed, &items); err != nil {
		return false
	}
	for _, it := range items {
		if it.Type == "tool_result" {
			return false
		}
	}
	for _, it := range items {
		if it.Type == "text" {
			return true
		}
	}
	return false
}
