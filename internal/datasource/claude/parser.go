package claude

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/breinzhang/tokusage/internal/domain"
	"github.com/breinzhang/tokusage/internal/platform"
)

const maxJSONLLineSize = 16 * 1024 * 1024

type ParseWarning struct {
	File    string
	Line    int64
	Message string
}

func ParseJSONLFile(path string) ([]domain.UsageEvent, []ParseWarning, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	sourceFile := platform.NormalizePathForStorage(path)
	var events []domain.UsageEvent
	var warnings []ParseWarning
	seenMessageIDs := map[string]bool{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), maxJSONLLineSize)
	lineNo := int64(0)
	for scanner.Scan() {
		lineNo++
		line := scanner.Bytes()
		var record rawRecord
		if err := json.Unmarshal(line, &record); err != nil {
			warnings = append(warnings, ParseWarning{File: path, Line: lineNo, Message: "malformed JSONL"})
			continue
		}
		if record.Type != "assistant" || record.Message.Usage == nil {
			continue
		}
		timestamp, err := time.Parse(time.RFC3339Nano, record.Timestamp)
		if err != nil {
			warnings = append(warnings, ParseWarning{File: path, Line: lineNo, Message: "missing or invalid timestamp"})
			continue
		}
		messageID := record.Message.ID
		if messageID == "" {
			messageID = fallbackMessageID(sourceFile, lineNo, record)
		}
		if seenMessageIDs[messageID] {
			continue
		}
		seenMessageIDs[messageID] = true

		model := record.Message.Model
		if model == "" {
			model = "unknown"
		}
		project := ResolveProject(record.CWD, transcriptProjectDirName(path))
		event := domain.UsageEvent{
			Agent:         "claude-code",
			ProjectID:     project.ID,
			ProjectName:   project.Name,
			ProjectPath:   project.PathNorm,
			SessionID:     sessionID(record.SessionID, path),
			MessageID:     messageID,
			AgentID:       record.AgentID,
			ParentAgentID: record.ParentAgentID,
			IsSubagent:    record.AgentID != "" || containsSubagentDir(path),
			Model:         model,
			Timestamp:     timestamp,
			Tokens:        tokensFromUsage(*record.Message.Usage, &warnings, path, lineNo),
			SourceFile:    sourceFile,
			RawLineNo:     lineNo,
		}
		event.EventHash = eventHash(event)
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, warnings, err
	}
	return events, warnings, nil
}

type rawRecord struct {
	Type          string     `json:"type"`
	SessionID     string     `json:"sessionId"`
	Timestamp     string     `json:"timestamp"`
	CWD           string     `json:"cwd"`
	AgentID       string     `json:"agentId"`
	ParentAgentID string     `json:"parentAgentId"`
	Message       rawMessage `json:"message"`
}

type rawMessage struct {
	ID    string    `json:"id"`
	Model string    `json:"model"`
	Usage *rawUsage `json:"usage"`
}

type rawUsage struct {
	InputTokens              int64             `json:"input_tokens"`
	OutputTokens             int64             `json:"output_tokens"`
	CacheCreationInputTokens int64             `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64             `json:"cache_read_input_tokens"`
	CacheCreation            *rawCacheCreation `json:"cache_creation"`
}

type rawCacheCreation struct {
	Ephemeral5mInputTokens int64 `json:"ephemeral_5m_input_tokens"`
	Ephemeral1hInputTokens int64 `json:"ephemeral_1h_input_tokens"`
}

func tokensFromUsage(usage rawUsage, warnings *[]ParseWarning, file string, line int64) domain.TokenSummary {
	tokens := domain.TokenSummary{
		StandardInputTokens: usage.InputTokens,
		OutputTokens:        usage.OutputTokens,
		CacheReadTokens:     usage.CacheReadInputTokens,
	}
	if usage.CacheCreation != nil {
		tokens.CacheWrite5mTokens = usage.CacheCreation.Ephemeral5mInputTokens
		tokens.CacheWrite1hTokens = usage.CacheCreation.Ephemeral1hInputTokens
	}
	knownCreation := tokens.CacheWrite5mTokens + tokens.CacheWrite1hTokens
	if usage.CacheCreationInputTokens > knownCreation {
		tokens.CacheWrite5mTokens += usage.CacheCreationInputTokens - knownCreation
		*warnings = append(*warnings, ParseWarning{File: file, Line: line, Message: "cache creation tokens lacked full tier detail; assigned remainder to 5m"})
	}
	return tokens
}

func sessionID(value string, path string) string {
	if value != "" {
		return value
	}
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func transcriptProjectDirName(path string) string {
	fallback := filepath.Base(filepath.Dir(path))
	parts := strings.Split(filepath.ToSlash(filepath.Clean(path)), "/")
	for i, part := range parts {
		if part == "subagents" && i > 0 {
			return parts[i-1]
		}
	}
	return fallback
}

func fallbackMessageID(path string, lineNo int64, record rawRecord) string {
	var usage rawUsage
	if record.Message.Usage != nil {
		usage = *record.Message.Usage
	}
	var cacheCreation rawCacheCreation
	if usage.CacheCreation != nil {
		cacheCreation = *usage.CacheCreation
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf(
		"%s:%d:%s:%s:%d:%d:%d:%d:%d:%d",
		path,
		lineNo,
		record.Timestamp,
		record.Message.Model,
		usage.InputTokens,
		usage.OutputTokens,
		usage.CacheCreationInputTokens,
		usage.CacheReadInputTokens,
		cacheCreation.Ephemeral5mInputTokens,
		cacheCreation.Ephemeral1hInputTokens,
	)))
	return "fallback-" + hex.EncodeToString(sum[:])
}

func containsSubagentDir(path string) bool {
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part == "subagents" {
			return true
		}
	}
	return false
}

func eventHash(event domain.UsageEvent) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf(
		"%s:%s:%s:%s:%d:%d:%d:%d:%d",
		event.Agent,
		event.SourceFile,
		event.SessionID,
		event.MessageID,
		event.Tokens.StandardInputTokens,
		event.Tokens.OutputTokens,
		event.Tokens.CacheWrite5mTokens,
		event.Tokens.CacheWrite1hTokens,
		event.Tokens.CacheReadTokens,
	)))
	return hex.EncodeToString(sum[:])
}
