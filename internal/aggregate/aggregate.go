package aggregate

import "github.com/breinzhang/tokusage/internal/cache"

func ByModel(rollups []cache.DailyRollup) map[string]Summary {
	result := map[string]Summary{}
	for _, rollup := range rollups {
		current := result[rollup.Model]
		current.Label = rollup.Model
		current.Tokens = current.Tokens.Add(rollup.Tokens)
		result[rollup.Model] = current
	}
	return result
}

func ByProject(rollups []cache.DailyRollup) map[string]Summary {
	result := map[string]Summary{}
	for _, rollup := range rollups {
		current := result[rollup.ProjectID]
		current.Label = rollup.ProjectID
		current.Tokens = current.Tokens.Add(rollup.Tokens)
		result[rollup.ProjectID] = current
	}
	return result
}

func ByDate(rollups []cache.DailyRollup) map[string]Summary {
	result := map[string]Summary{}
	for _, rollup := range rollups {
		current := result[rollup.Date]
		current.Label = rollup.Date
		current.Tokens = current.Tokens.Add(rollup.Tokens)
		result[rollup.Date] = current
	}
	return result
}
