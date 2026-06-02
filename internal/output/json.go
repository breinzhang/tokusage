package output

import (
	"encoding/json"

	"github.com/breinzhang/tokusage/internal/domain"
)

type Report struct {
	Agent        string   `json:"agent"`
	GeneratedAt  string   `json:"generated_at,omitempty"`
	Timezone     string   `json:"timezone"`
	From         string   `json:"from"`
	To           string   `json:"to"`
	LabelHeader  string   `json:"label_header,omitempty"`
	Estimated    bool     `json:"estimated"`
	Partial      bool     `json:"partial"`
	PricingHash  string   `json:"pricing_hash"`
	Warnings     []string `json:"warnings"`
	Buckets      []Bucket `json:"buckets"`
	ColorEnabled bool     `json:"-"`
}

type Bucket struct {
	Label         string              `json:"label"`
	Tokens        domain.TokenSummary `json:"tokens"`
	TotalTokens   int64               `json:"total_tokens"`
	EstimatedCost string              `json:"estimated_cost_usd,omitempty"`
	Partial       bool                `json:"partial,omitempty"`
}

func RenderJSON(report Report) ([]byte, error) {
	for i := range report.Buckets {
		report.Buckets[i].TotalTokens = report.Buckets[i].Tokens.TotalTokens()
	}
	return json.MarshalIndent(report, "", "  ")
}
