package timeutil

import (
	"errors"
	"strings"
	"time"
)

var (
	errInvalidFormat = errors.New("Invalid format")
	dateLayout       = "2006-01-2"
	timeLayout       = "15:04"
	layout           = "2006-01-02 15:04"
	// FIX: handle timezone
	loc = time.Local
)

func ParseAndValidateTimestamp(timestampStr string) (time.Time, error) {
	var timestamp time.Time
	timestampStr = strings.TrimSpace(timestampStr)

	if strings.Contains(timestampStr, " ") {
		datetime := strings.Split(timestampStr, " ")
		if len(datetime) != 2 {
			return timestamp, errInvalidFormat
		}

		dateStr, err := ValidateDate(datetime[0], loc)
		if err != nil {
			return timestamp, errInvalidFormat
		}
		_, err = time.Parse(timeLayout, datetime[1])
		if err != nil {
			return timestamp, errInvalidFormat
		}

		timestampStr := dateStr + " " + datetime[1]
		timestamp, err = time.ParseInLocation(layout, timestampStr, loc)
		if err != nil {
			return timestamp, errInvalidFormat
		}
	} else {
		dateStr, err := ValidateDate(timestampStr, loc)
		if err != nil {
			return timestamp, errInvalidFormat
		}

		timestamp, err = time.ParseInLocation(dateLayout, dateStr, loc)
		if err != nil {
			return timestamp, nil
		}
	}
	return timestamp, nil
}

func ValidateDate(dateStr string, loc *time.Location) (string, error) {
	date, err := time.ParseInLocation(dateLayout, dateStr, loc)
	if err != nil {
		dateStr = strings.ToLower(dateStr)
		// dateStr = strings.TrimSpace(dateStr)
		today := time.Now()
		switch dateStr {
		case "today":
			date = today
		case "tommorow":
			date = date.AddDate(0, 0, 1)
		}
	}
	return date.Format(dateLayout), nil
}
