package timeutil

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	dateLayout = "2006-01-02"
	timeLayout = "15:04"
	layout     = "2006-01-02 15:04"

	inNUnitRegex = regexp.MustCompile(`^in (\d+) (day|week|month|year)s?$`)

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
	weekdayStr = strings.TrimSpace(weekdayStr)
	weekdayStr = strings.ToLower(weekdayStr)
	if wd, ok := weekdayMap[weekdayStr]; ok {
		return wd, nil
	}
	return time.Monday, errors.New("error parsing weekday: invalid format")
}

func NearestWeekday(startDate time.Time, tartgetWeekday time.Weekday) time.Time {
	daysToAdd := (int(tartgetWeekday) - int(startDate.Weekday()))
	if daysToAdd < 0 {
		daysToAdd += 7
	}

	return startDate.AddDate(0, 0, daysToAdd)
}

func SeparateDateAndTime(input string) (dateStr, timeStr string) {
	timeRegex := regexp.MustCompile(`(\d{1,2}):(\d{2})`)
	timeMatches := timeRegex.FindStringSubmatch(input)

	if len(timeMatches) > 0 {
		timeStr = timeMatches[0]

		dateStr = strings.Replace(input, timeStr, "", 1)
		dateStr = strings.TrimSpace(dateStr)
	} else {
		dateStr = input
		timeStr = ""
	}
	return dateStr, timeStr
}

func ParseAndValidateTimestamp(datetimeStr string) (time.Time, error) {
	datetimeStr = strings.TrimSpace(datetimeStr)
	datetimeStr = strings.ToLower(datetimeStr)
	dateStr, timeStr := SeparateDateAndTime(datetimeStr)

	timestampDate, err := ParseAndValidateDate(dateStr)
	if err != nil {
		return time.Time{}, err
	}

	if timeStr != "" {
		timestampTime, err := time.ParseInLocation(timeLayout, timeStr, loc)
		if err != nil {
			return time.Time{}, errors.New("invaild time format")
		}
		return time.Date(timestampDate.Year(), timestampDate.Month(), timestampDate.Day(), timestampTime.Hour(), timestampTime.Minute(), 0, 0, loc), nil
	}

	return timestampDate, nil
}

func ParseAndValidateDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	dateStr = strings.ToLower(dateStr)
	now := time.Now()
	date, _ := time.ParseInLocation(layout, dateStr, loc)
	if date.IsZero() {
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		switch dateStr {
		case "today":
			return today, nil
		case "tommorow":
			return today.AddDate(0, 0, 1), nil
		case "next week":
			return today.AddDate(0, 0, 7), nil
		case "next month":
			return today.AddDate(0, 1, 0), nil
		}
		if strings.HasPrefix(dateStr, "next ") {
			parts := strings.Fields(dateStr)
			if len(parts) == 2 {
				targetWd, err := ParseWeekDay(parts[1])
				if err != nil {
					return time.Time{}, err
				}
				wd := NearestWeekday(today, targetWd)
				return wd, nil
			}
		}
		if matches := inNUnitRegex.FindStringSubmatch(dateStr); len(matches) > 0 {
			n, err := strconv.Atoi(matches[1])
			if err != nil {
				return time.Time{}, errors.New("failed to parse number")
			}
			unit := matches[2]
			switch unit {
			case "days":
				return today.AddDate(0, 0, n), nil
			case "weeks":
				return today.AddDate(0, 0, n*7), nil
			case "months":
				return today.AddDate(0, n, 0), nil
			case "years":
				return today.AddDate(n, 0, 0), nil
			}
		}
		return time.Time{}, errors.New("unable to parse date string")
	}
	return date, nil
}
