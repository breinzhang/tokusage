package aggregate

import "github.com/breinzhang/tokusage/internal/domain"

type Summary struct {
	Label    string
	Tokens   domain.TokenSummary
	Models   map[string]Summary
	Projects map[string]Summary
}
