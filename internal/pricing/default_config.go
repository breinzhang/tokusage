package pricing

import (
	"os"
	"path/filepath"
	"strings"
)

const defaultConfig = `
# tokusage pricing config.
# Prices are USD per 1M tokens. Patterns use Go filepath-style matching.

# Anthropic
[[models]]
pattern = "claude-opus-4-8*"
standard_input_per_mtok = "5.00"
cache_write_5m_per_mtok = "6.25"
cache_write_1h_per_mtok = "10.00"
cache_read_per_mtok = "0.50"
output_per_mtok = "25.00"

[[models]]
pattern = "claude-opus-4.8*"
standard_input_per_mtok = "5.00"
cache_write_5m_per_mtok = "6.25"
cache_write_1h_per_mtok = "10.00"
cache_read_per_mtok = "0.50"
output_per_mtok = "25.00"

[[models]]
pattern = "claude-opus-4-7*"
standard_input_per_mtok = "5.00"
cache_write_5m_per_mtok = "6.25"
cache_write_1h_per_mtok = "10.00"
cache_read_per_mtok = "0.50"
output_per_mtok = "25.00"

[[models]]
pattern = "claude-opus-4.7*"
standard_input_per_mtok = "5.00"
cache_write_5m_per_mtok = "6.25"
cache_write_1h_per_mtok = "10.00"
cache_read_per_mtok = "0.50"
output_per_mtok = "25.00"

[[models]]
pattern = "claude-opus-4-6*"
standard_input_per_mtok = "5.00"
cache_write_5m_per_mtok = "6.25"
cache_write_1h_per_mtok = "10.00"
cache_read_per_mtok = "0.50"
output_per_mtok = "25.00"

[[models]]
pattern = "claude-opus-4.6*"
standard_input_per_mtok = "5.00"
cache_write_5m_per_mtok = "6.25"
cache_write_1h_per_mtok = "10.00"
cache_read_per_mtok = "0.50"
output_per_mtok = "25.00"

[[models]]
pattern = "claude-opus-4-5*"
standard_input_per_mtok = "5.00"
cache_write_5m_per_mtok = "6.25"
cache_write_1h_per_mtok = "10.00"
cache_read_per_mtok = "0.50"
output_per_mtok = "25.00"

[[models]]
pattern = "claude-opus-4.5*"
standard_input_per_mtok = "5.00"
cache_write_5m_per_mtok = "6.25"
cache_write_1h_per_mtok = "10.00"
cache_read_per_mtok = "0.50"
output_per_mtok = "25.00"

[[models]]
pattern = "claude-opus-4-1*"
standard_input_per_mtok = "15.00"
cache_write_5m_per_mtok = "18.75"
cache_write_1h_per_mtok = "30.00"
cache_read_per_mtok = "1.50"
output_per_mtok = "75.00"

[[models]]
pattern = "claude-opus-4.1*"
standard_input_per_mtok = "15.00"
cache_write_5m_per_mtok = "18.75"
cache_write_1h_per_mtok = "30.00"
cache_read_per_mtok = "1.50"
output_per_mtok = "75.00"

[[models]]
pattern = "claude-opus-4*"
standard_input_per_mtok = "15.00"
cache_write_5m_per_mtok = "18.75"
cache_write_1h_per_mtok = "30.00"
cache_read_per_mtok = "1.50"
output_per_mtok = "75.00"

[[models]]
pattern = "claude-sonnet-4*"
standard_input_per_mtok = "3.00"
cache_write_5m_per_mtok = "3.75"
cache_write_1h_per_mtok = "6.00"
cache_read_per_mtok = "0.30"
output_per_mtok = "15.00"

[[models]]
pattern = "claude-haiku-4.5*"
standard_input_per_mtok = "1.00"
cache_write_5m_per_mtok = "1.25"
cache_write_1h_per_mtok = "2.00"
cache_read_per_mtok = "0.10"
output_per_mtok = "5.00"

[[models]]
pattern = "claude-haiku-4*"
standard_input_per_mtok = "1.00"
cache_write_5m_per_mtok = "1.25"
cache_write_1h_per_mtok = "2.00"
cache_read_per_mtok = "0.10"
output_per_mtok = "5.00"

[[models]]
pattern = "claude-3-5-haiku*"
standard_input_per_mtok = "0.80"
cache_write_5m_per_mtok = "1.00"
cache_write_1h_per_mtok = "1.60"
cache_read_per_mtok = "0.08"
output_per_mtok = "4.00"

# Kimi/Moonshot
[[models]]
pattern = "kimi-k2.7-code*"
standard_input_per_mtok = "0.95"
cache_write_5m_per_mtok = "0.95"
cache_write_1h_per_mtok = "0.95"
cache_read_per_mtok = "0.19"
output_per_mtok = "4.00"

[[models]]
pattern = "kimi-k2.6*"
standard_input_per_mtok = "0.95"
cache_write_5m_per_mtok = "0.95"
cache_write_1h_per_mtok = "0.95"
cache_read_per_mtok = "0.16"
output_per_mtok = "4.00"

[[models]]
pattern = "kimi-k2.5*"
standard_input_per_mtok = "0.60"
cache_write_5m_per_mtok = "0.60"
cache_write_1h_per_mtok = "0.60"
cache_read_per_mtok = "0.10"
output_per_mtok = "3.00"

# GLM/Z.AI
[[models]]
pattern = "glm-5.2*"
standard_input_per_mtok = "1.40"
cache_write_5m_per_mtok = "1.40"
cache_write_1h_per_mtok = "1.40"
cache_read_per_mtok = "0.26"
output_per_mtok = "4.40"

[[models]]
pattern = "glm-5.1*"
standard_input_per_mtok = "1.40"
cache_write_5m_per_mtok = "1.40"
cache_write_1h_per_mtok = "1.40"
cache_read_per_mtok = "0.26"
output_per_mtok = "4.40"

[[models]]
pattern = "glm-5-turbo*"
standard_input_per_mtok = "1.20"
cache_write_5m_per_mtok = "1.20"
cache_write_1h_per_mtok = "1.20"
cache_read_per_mtok = "0.24"
output_per_mtok = "4.00"

[[models]]
pattern = "glm-5*"
standard_input_per_mtok = "1.00"
cache_write_5m_per_mtok = "1.00"
cache_write_1h_per_mtok = "1.00"
cache_read_per_mtok = "0.20"
output_per_mtok = "3.20"

[[models]]
pattern = "glm-4.7-flashx*"
standard_input_per_mtok = "0.07"
cache_write_5m_per_mtok = "0.07"
cache_write_1h_per_mtok = "0.07"
cache_read_per_mtok = "0.01"
output_per_mtok = "0.40"

[[models]]
pattern = "glm-4.7-flash*"
standard_input_per_mtok = "0.00"
cache_write_5m_per_mtok = "0.00"
cache_write_1h_per_mtok = "0.00"
cache_read_per_mtok = "0.00"
output_per_mtok = "0.00"

[[models]]
pattern = "glm-4.7*"
standard_input_per_mtok = "0.60"
cache_write_5m_per_mtok = "0.60"
cache_write_1h_per_mtok = "0.60"
cache_read_per_mtok = "0.11"
output_per_mtok = "2.20"

[[models]]
pattern = "glm-4.6*"
standard_input_per_mtok = "0.60"
cache_write_5m_per_mtok = "0.60"
cache_write_1h_per_mtok = "0.60"
cache_read_per_mtok = "0.11"
output_per_mtok = "2.20"

[[models]]
pattern = "glm-4.5-airx*"
standard_input_per_mtok = "1.10"
cache_write_5m_per_mtok = "1.10"
cache_write_1h_per_mtok = "1.10"
cache_read_per_mtok = "0.22"
output_per_mtok = "4.50"

[[models]]
pattern = "glm-4.5-air*"
standard_input_per_mtok = "0.20"
cache_write_5m_per_mtok = "0.20"
cache_write_1h_per_mtok = "0.20"
cache_read_per_mtok = "0.03"
output_per_mtok = "1.10"

[[models]]
pattern = "glm-4.5-x*"
standard_input_per_mtok = "2.20"
cache_write_5m_per_mtok = "2.20"
cache_write_1h_per_mtok = "2.20"
cache_read_per_mtok = "0.45"
output_per_mtok = "8.90"

[[models]]
pattern = "glm-4.5-flash*"
standard_input_per_mtok = "0.00"
cache_write_5m_per_mtok = "0.00"
cache_write_1h_per_mtok = "0.00"
cache_read_per_mtok = "0.00"
output_per_mtok = "0.00"

[[models]]
pattern = "glm-4.5*"
standard_input_per_mtok = "0.60"
cache_write_5m_per_mtok = "0.60"
cache_write_1h_per_mtok = "0.60"
cache_read_per_mtok = "0.11"
output_per_mtok = "2.20"

[[models]]
pattern = "glm-4-32b-0414-128k*"
standard_input_per_mtok = "0.10"
cache_write_5m_per_mtok = "0.10"
cache_write_1h_per_mtok = "0.10"
cache_read_per_mtok = "0.10"
output_per_mtok = "0.10"

# DeepSeek
[[models]]
pattern = "deepseek-v4-pro*"
standard_input_per_mtok = "0.435"
cache_write_5m_per_mtok = "0.435"
cache_write_1h_per_mtok = "0.435"
cache_read_per_mtok = "0.003625"
output_per_mtok = "0.87"

[[models]]
pattern = "deepseek-v4-flash*"
standard_input_per_mtok = "0.14"
cache_write_5m_per_mtok = "0.14"
cache_write_1h_per_mtok = "0.14"
cache_read_per_mtok = "0.0028"
output_per_mtok = "0.28"

[[models]]
pattern = "deepseek-chat"
standard_input_per_mtok = "0.14"
cache_write_5m_per_mtok = "0.14"
cache_write_1h_per_mtok = "0.14"
cache_read_per_mtok = "0.0028"
output_per_mtok = "0.28"

[[models]]
pattern = "deepseek-reasoner"
standard_input_per_mtok = "0.14"
cache_write_5m_per_mtok = "0.14"
cache_write_1h_per_mtok = "0.14"
cache_read_per_mtok = "0.0028"
output_per_mtok = "0.28"

# MiniMax
[[models]]
pattern = "minimax-m3*"
standard_input_per_mtok = "0.30"
cache_write_5m_per_mtok = "0.30"
cache_write_1h_per_mtok = "0.30"
cache_read_per_mtok = "0.06"
output_per_mtok = "1.20"

[[models]]
pattern = "minimax-m2.7-highspeed*"
standard_input_per_mtok = "0.60"
cache_write_5m_per_mtok = "0.375"
cache_write_1h_per_mtok = "0.375"
cache_read_per_mtok = "0.06"
output_per_mtok = "2.40"

[[models]]
pattern = "minimax-m2.7*"
standard_input_per_mtok = "0.30"
cache_write_5m_per_mtok = "0.375"
cache_write_1h_per_mtok = "0.375"
cache_read_per_mtok = "0.06"
output_per_mtok = "1.20"

[[models]]
pattern = "minimax-m2.5-highspeed*"
standard_input_per_mtok = "0.60"
cache_write_5m_per_mtok = "0.375"
cache_write_1h_per_mtok = "0.375"
cache_read_per_mtok = "0.03"
output_per_mtok = "2.40"

[[models]]
pattern = "minimax-m2.5*"
standard_input_per_mtok = "0.30"
cache_write_5m_per_mtok = "0.375"
cache_write_1h_per_mtok = "0.375"
cache_read_per_mtok = "0.03"
output_per_mtok = "1.20"

# Fallback
[[models]]
pattern = "*"
standard_input_per_mtok = "3.00"
cache_write_5m_per_mtok = "3.75"
cache_write_1h_per_mtok = "6.00"
cache_read_per_mtok = "0.30"
output_per_mtok = "15.00"
`

func EnsureDefaultConfig(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	_, err = file.WriteString(strings.TrimSpace(defaultConfig) + "\n")
	return err
}
