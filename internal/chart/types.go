package chart

import "github.com/shopspring/decimal"

type Metric string

const (
	MetricTokens Metric = "tokens"
	MetricCost   Metric = "cost"
)

type SplitBy string

const (
	SplitNone    SplitBy = "none"
	SplitModel   SplitBy = "model"
	SplitProject SplitBy = "project"
)

type ColorMode string

const (
	ColorAuto   ColorMode = "auto"
	ColorNever  ColorMode = "never"
	ColorAlways ColorMode = "always"
)

type Options struct {
	GroupBy      string
	Metric       Metric
	SplitBy      SplitBy
	Top          int
	Width        int
	ASCII        bool
	Color        ColorMode
	ColorEnabled bool
	From         string
	To           string
}

type Chart struct {
	Metric  Metric
	GroupBy string
	SplitBy SplitBy
	From    string
	To      string
	Buckets []Bucket
	Legend  []LegendItem
}

type Bucket struct {
	Label    string
	Value    decimal.Decimal
	Segments []Segment
}

type Segment struct {
	Label string
	Value decimal.Decimal
}

type LegendItem struct {
	Key   string
	Label string
	Value decimal.Decimal
	Share decimal.Decimal
}
