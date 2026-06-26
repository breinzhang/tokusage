package pricing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigPricing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pricing.toml")
	content := []byte(`
[[models]]
pattern = "glm-5.1"
standard_input_per_mtok = "1.00"
cache_write_5m_per_mtok = "1.25"
cache_write_1h_per_mtok = "2.00"
cache_read_per_mtok = "0.10"
output_per_mtok = "5.00"
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	prices, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(prices) != 1 {
		t.Fatalf("len(prices) = %d, want 1", len(prices))
	}
	if prices[0].Pattern != "glm-5.1" {
		t.Fatalf("Pattern = %q", prices[0].Pattern)
	}
	if prices[0].CacheWrite1hPerMTok.StringFixed(2) != "2.00" {
		t.Fatalf("CacheWrite1hPerMTok = %s", prices[0].CacheWrite1hPerMTok)
	}
}

func TestLoadConfigRejectsInvalidDecimal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pricing.toml")
	content := []byte(`
[[models]]
pattern = "bad"
standard_input_per_mtok = "not-money"
cache_write_5m_per_mtok = "1.25"
cache_write_1h_per_mtok = "2.00"
cache_read_per_mtok = "0.10"
output_per_mtok = "5.00"
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("LoadConfig error = nil, want error")
	}
}

func TestEnsureDefaultConfigWritesLoadableCommonModelPrices(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".tokusage", "pricing.toml")

	if err := EnsureDefaultConfig(path); err != nil {
		t.Fatal(err)
	}
	prices, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	provider := NewStaticProvider(prices)
	for _, model := range []string{"claude-sonnet-4.5", "glm-5.1", "kimi-k2.7-code", "deepseek-chat", "minimax-m3"} {
		if _, ok := provider.PriceFor(model); !ok {
			t.Fatalf("default config missing price for %s", model)
		}
	}

	glm, ok := provider.PriceFor("glm-5.1")
	if !ok {
		t.Fatal("glm-5.1 price not found")
	}
	if glm.StandardInputPerMTok.StringFixed(2) != "1.40" || glm.OutputPerMTok.StringFixed(2) != "4.40" {
		t.Fatalf("glm price = input %s output %s, want CC-Link defaults", glm.StandardInputPerMTok, glm.OutputPerMTok)
	}
}

func TestEnsureDefaultConfigDoesNotOverwriteExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".tokusage", "pricing.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	original := []byte("# custom pricing\n")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureDefaultConfig(path); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("existing config was overwritten:\n%s", got)
	}
}
