package pricing

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/shopspring/decimal"
)

type configFile struct {
	Models []configModel `toml:"models"`
}

type configModel struct {
	Pattern                  string `toml:"pattern"`
	StandardInputPerMTokText string `toml:"standard_input_per_mtok"`
	CacheWrite5mPerMTokText  string `toml:"cache_write_5m_per_mtok"`
	CacheWrite1hPerMTokText  string `toml:"cache_write_1h_per_mtok"`
	CacheReadPerMTokText     string `toml:"cache_read_per_mtok"`
	OutputPerMTokText        string `toml:"output_per_mtok"`
}

func LoadConfig(path string) ([]ModelPrice, error) {
	var cfg configFile
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}

	prices := make([]ModelPrice, 0, len(cfg.Models))
	for i, model := range cfg.Models {
		price, err := model.toPrice(i)
		if err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}
	return prices, nil
}

func (m configModel) toPrice(index int) (ModelPrice, error) {
	if m.Pattern == "" {
		return ModelPrice{}, fmt.Errorf("models[%d].pattern is required", index)
	}
	standardInput, err := parsePrice(index, "standard_input_per_mtok", m.StandardInputPerMTokText)
	if err != nil {
		return ModelPrice{}, err
	}
	cacheWrite5m, err := parsePrice(index, "cache_write_5m_per_mtok", m.CacheWrite5mPerMTokText)
	if err != nil {
		return ModelPrice{}, err
	}
	cacheWrite1h, err := parsePrice(index, "cache_write_1h_per_mtok", m.CacheWrite1hPerMTokText)
	if err != nil {
		return ModelPrice{}, err
	}
	cacheRead, err := parsePrice(index, "cache_read_per_mtok", m.CacheReadPerMTokText)
	if err != nil {
		return ModelPrice{}, err
	}
	output, err := parsePrice(index, "output_per_mtok", m.OutputPerMTokText)
	if err != nil {
		return ModelPrice{}, err
	}
	return ModelPrice{
		Pattern:              m.Pattern,
		StandardInputPerMTok: standardInput,
		CacheWrite5mPerMTok:  cacheWrite5m,
		CacheWrite1hPerMTok:  cacheWrite1h,
		CacheReadPerMTok:     cacheRead,
		OutputPerMTok:        output,
	}, nil
}

func parsePrice(index int, field string, value string) (decimal.Decimal, error) {
	if value == "" {
		return decimal.Decimal{}, fmt.Errorf("models[%d].%s is required", index, field)
	}
	result, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("models[%d].%s: %w", index, field, err)
	}
	return result, nil
}
