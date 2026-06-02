package domain

import "time"

type UsageEvent struct {
	Agent       string
	ProjectID   string
	ProjectName string
	ProjectPath string

	SessionID     string
	MessageID     string
	AgentID       string
	ParentAgentID string
	IsSubagent    bool

	Model     string
	Timestamp time.Time

	Tokens TokenSummary

	SourceFile string
	RawLineNo  int64
	EventHash  string
}
