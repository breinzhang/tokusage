package pricing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"github.com/shopspring/decimal"
)

type ModelPrice struct {
	Pattern              string
	StandardInputPerMTok decimal.Decimal
	CacheWrite5mPerMTok  decimal.Decimal
	CacheWrite1hPerMTok  decimal.Decimal
	CacheReadPerMTok     decimal.Decimal
	OutputPerMTok        decimal.Decimal
}

type Provider interface {
	PriceFor(model string) (ModelPrice, bool)
	Hash() string
}

type StaticProvider struct {
	prices []ModelPrice
	hash   string
}

func NewStaticProvider(prices []ModelPrice) StaticProvider {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%+v", prices)))
	return StaticProvider{prices: prices, hash: "sha256:" + hex.EncodeToString(sum[:])}
}

func (p StaticProvider) PriceFor(model string) (ModelPrice, bool) {
	for _, price := range p.prices {
		matched, _ := filepath.Match(price.Pattern, model)
		if matched || price.Pattern == model {
			return price, true
		}
	}
	return ModelPrice{}, false
}

func (p StaticProvider) Hash() string {
	return p.hash
}
