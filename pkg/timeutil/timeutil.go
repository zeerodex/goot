package timeutil

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	errInvalidFormat = errors.New("Invalid format")
	dateLayout       = "2006-01-02"
	timeLayout       = "15:04"
	layout           = "2006-01-02 15:04"

	weekdayMap = map[string]time.Weekday{
		"sunday":    time.Sunday,
		"sun":       time.Sunday,
		"monday":    time.Monday,
		"mon":       time.Monday,
		"tuesday":   time.Tuesday,
		"tue":       time.Tuesday,
		"wednesday": time.Wednesday,
		"wed":       time.Wednesday,
		"thursday":  time.Thursday,
		"thu":       time.Thursday,
		"friday":    time.Friday,
		"fri":       time.Friday,
		"saturday":  time.Saturday,
		"sat":       time.Saturday,
	}

	// FIX: handle timezone
	loc = time.Local
)

func ParseWeekDay(weekdayStr string) (time.Weekday, error) {
	weekdayStr = strings.ToLower(weekdayStr)
	if wd, ok := weekdayMap[weekdayStr]; ok {
		return wd, nil
	}
	return time.Monday, errInvalidFormat
}

func NearestWeekday(currentWeekday, tartgetWeekday time.Weekday) time.Time {
	daysToAdd := (int(tartgetWeekday) - int(currentWeekday))
	if daysToAdd < 0 {
		daysToAdd += 7
	}

	return time.Now().AddDate(0, 0, daysToAdd)
}

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
			fmt.Println(err)
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
	today := time.Now()
	date, err := time.ParseInLocation(dateLayout, dateStr, loc)
	if err != nil {
		dateStr = strings.ToLower(dateStr)
		// dateStr = strings.TrimSpace(dateStr)
		switch dateStr {
		case "today":
			date = today
		case "tommorow":
			date = today.AddDate(0, 0, 1)
		}
		if strings.HasPrefix("next", dateStr) || strings.HasPrefix("this", dateStr) {
			dateSlice := strings.Split(dateStr, " ")
			if len(dateSlice) == 2 {
				after := dateSlice[0]
				period := dateSlice[1]
				weekday, err := ParseWeekDay(period)
				if err != nil {
					if period == "week" {
						today = today.AddDate(0, 0, 7)
					} else {
						return "", errInvalidFormat
					}
				}

				date = NearestWeekday(today.Weekday(), weekday)
				// HACK:
				if after == "next" {
					date = date.AddDate(0, 0, 7)
				}
			} else {
				return "", errInvalidFormat
			}
		} else {
			weekday, err := ParseWeekDay(dateStr)
			if err != nil {
				if dateStr == "week" {
					today = today.AddDate(0, 0, 7)
				} else {
					return "", errInvalidFormat
				}
			}

			date = NearestWeekday(today.Weekday(), weekday)
		}
	}
	if date.Format(dateLayout) != "0001-01-01" {
		return date.Format(dateLayout), nil
	} else {
		return "", errInvalidFormat
	}
}
