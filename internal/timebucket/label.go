package timebucket

import (
	"fmt"
	"time"
)

func Label(groupBy string, dateText string) (string, error) {
	date, err := time.Parse(time.DateOnly, dateText)
	if err != nil {
		return "", fmt.Errorf("invalid date %q: %w", dateText, err)
	}

	switch groupBy {
	case "day":
		return date.Format(time.DateOnly), nil
	case "month":
		return date.Format("2006-01"), nil
	case "week":
		return fmt.Sprintf("%s W%d", date.Format("2006-01"), weekOfMonth(date)), nil
	case "year":
		return date.Format("2006"), nil
	default:
		return "", fmt.Errorf("unsupported group-by %q", groupBy)
	}
}

func weekOfMonth(date time.Time) int {
	first := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	mondayOffset := (int(first.Weekday()) + 6) % 7
	return (date.Day()+mondayOffset-1)/7 + 1
}
