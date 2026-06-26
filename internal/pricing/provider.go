package pricing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

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

type OverlayProvider struct {
	primary  Provider
	fallback Provider
	hash     string
}

func NewStaticProvider(prices []ModelPrice) StaticProvider {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%+v", prices)))
	return StaticProvider{prices: prices, hash: "sha256:" + hex.EncodeToString(sum[:])}
}

func NewOverlayProvider(primary Provider, fallback Provider) OverlayProvider {
	sum := sha256.Sum256([]byte(primary.Hash() + "\n" + fallback.Hash()))
	return OverlayProvider{primary: primary, fallback: fallback, hash: "sha256:" + hex.EncodeToString(sum[:])}
}

func (p StaticProvider) PriceFor(model string) (ModelPrice, bool) {
	for _, price := range p.prices {
		pattern := strings.TrimSpace(price.Pattern)
		modelName := strings.TrimSpace(model)
		matched, _ := filepath.Match(strings.ToLower(pattern), strings.ToLower(modelName))
		if matched || strings.EqualFold(pattern, modelName) {
			return price, true
		}
	}
	return ModelPrice{}, false
}

func (p StaticProvider) Hash() string {
	return p.hash
}

func (p OverlayProvider) PriceFor(model string) (ModelPrice, bool) {
	if price, ok := p.primary.PriceFor(model); ok {
		return price, true
	}
	return p.fallback.PriceFor(model)
}

func (p OverlayProvider) Hash() string {
	return p.hash
}
