package formula_engine

import (
	"fmt"
	"time"
)

func funcNow(args []interface{}) (interface{}, error) {
	return time.Now(), nil
}

func funcToday(args []interface{}) (interface{}, error) {
	now := time.Now()
	// Return start of the day
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
}

func funcYear(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("YEAR requires 1 argument")
	}
	t, err := toTime(args[0])
	if err != nil {
		return nil, err
	}
	return float64(t.Year()), nil
}

func funcMonth(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("MONTH requires 1 argument")
	}
	t, err := toTime(args[0])
	if err != nil {
		return nil, err
	}
	return float64(t.Month()), nil
}

func toTime(v interface{}) (time.Time, error) {
	switch val := v.(type) {
	case time.Time:
		return val, nil
	case string:
		// Try parsing various formats
		// ISO 8601
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t, nil
		}
		if t, err := time.Parse("2006-01-02", val); err == nil {
			return t, nil
		}
		return time.Time{}, fmt.Errorf("cannot parse date: %s", val)
	default:
		return time.Time{}, fmt.Errorf("cannot convert type %T to date", val)
	}
}
