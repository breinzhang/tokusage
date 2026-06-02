package output

import (
	"encoding/json"
	"testing"

	"github.com/breinzhang/tokusage/internal/domain"
)

func TestRenderJSONUsesStableFields(t *testing.T) {
	report := Report{
		Agent:       "claude-code",
		Timezone:    "Asia/Shanghai",
		From:        "2026-05-01",
		To:          "2026-05-30",
		Estimated:   true,
		Partial:     true,
		PricingHash: "sha256:example",
		Buckets: []Bucket{{
			Label:  "2026-05-10",
			Tokens: domain.TokenSummary{StandardInputTokens: 10},
		}},
	}

	data, err := RenderJSON(report)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["agent"] != "claude-code" {
		t.Fatalf("agent = %v", decoded["agent"])
	}
	if decoded["partial"] != true {
		t.Fatalf("partial = %v", decoded["partial"])
	}
}
