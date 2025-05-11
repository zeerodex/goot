package timeutil

import (
	"testing"
	"time"
)

func TestParseWeekDay(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Weekday
		err      bool
	}{
		{"Sunday full", "sunday", time.Sunday, false},
		{"Sunday short", "sun", time.Sunday, false},
		{"Monday full", "monday", time.Monday, false},
		{"Monday short", "mon", time.Monday, false},
		{"Tuesday full", "tuesday", time.Tuesday, false},
		{"Tuesday short", "tue", time.Tuesday, false},
		{"Wednesday full", "wednesday", time.Wednesday, false},
		{"Wednesday short", "wed", time.Wednesday, false},
		{"Thursday full", "thursday", time.Thursday, false},
		{"Thursday short", "thu", time.Thursday, false},
		{"Friday full", "friday", time.Friday, false},
		{"Friday short", "fri", time.Friday, false},
		{"Saturday full", "saturday", time.Saturday, false},
		{"Saturday short", "sat", time.Saturday, false},
		{"With spacing", "  monday  ", time.Monday, false},
		{"Mixed case", "MoNdAy", time.Monday, false},
		{"Invalid day", "someday", time.Monday, true},
		{"Empty", "", time.Monday, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWeekDay(tt.input)
			if (err != nil) != tt.err {
				t.Errorf("ParseWeekDay(%q) error = %v, expected err %v", tt.input, err, tt.err)
				return
			}
			if !tt.err && got != tt.expected {
				t.Errorf("ParseWeekDay(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNearestWeekday(t *testing.T) {
	tests := []struct {
		name          string
		startDate     time.Time
		targetWeekday time.Weekday
		expectedDays  int
	}{
		{"Next day", time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), time.Tuesday, 1},
		{"Six days ahead", time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), time.Sunday, 6},
		{"Next week", time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), time.Monday, 7}, // This should actually go to next Monday (7 days)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NearestWeekday(tt.startDate, tt.targetWeekday)
			expectedDate := tt.startDate.AddDate(0, 0, tt.expectedDays)

			if !got.Equal(expectedDate) {
				t.Errorf("NearestWeekday(%v, %v) = %v, want %v",
					tt.startDate, tt.targetWeekday, got, expectedDate)
			}
		})
	}
}

func TestSeparateDateAndTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDate string
		wantTime string
	}{
		{"Date and time", "2023-01-02 15:30", "2023-01-02", "15:30"},
		{"Date only", "2023-01-02", "2023-01-02", ""},
		{"Natural language with time", "today 10:30", "today", "10:30"},
		{"Single digit hour", "tomorrow 9:45", "tomorrow", "9:45"},
		{"Extra spaces", "  next monday  14:00  ", "next monday", "14:00"},
		{"No separation", "2023-01-0215:30", "2023-01-02", "15:30"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDate, gotTime := SeparateDateAndTime(tt.input)
			if gotDate != tt.wantDate {
				t.Errorf("SeparateDateAndTime(%q) gotDate = %v, want %v", tt.input, gotDate, tt.wantDate)
			}
			if gotTime != tt.wantTime {
				t.Errorf("SeparateDateAndTime(%q) gotTime = %v, want %v", tt.input, gotTime, tt.wantTime)
			}
		})
	}
}

func TestParseAndValidateDate(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	tests := []struct {
		name     string
		input    string
		expected time.Time
		err      bool
	}{
		{"ISO date", "2023-05-15", time.Date(2023, 5, 15, 0, 0, 0, 0, loc), false},
		{"Today", "today", today, false},
		{"Tomorrow", "tommorow", today.AddDate(0, 0, 1), false}, // Note the typo in the original code
		{"Next week", "next week", today.AddDate(0, 0, 7), false},
		{"Next month", "next month", today.AddDate(0, 1, 0), false},
		{"Next Monday", "next monday", NearestWeekday(today, time.Monday), false},
		{"Next Friday", "next friday", NearestWeekday(today, time.Friday), false},
		{"In 3 days", "in 3 days", today.AddDate(0, 0, 3), false},
		{"In 2 weeks", "in 2 weeks", today.AddDate(0, 0, 14), false},
		{"In 1 month", "in 1 month", today.AddDate(0, 1, 0), false},
		{"In 5 years", "in 5 years", today.AddDate(5, 0, 0), false},
		{"Invalid input", "invalid date", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAndValidateDate(tt.input)
			if (err != nil) != tt.err {
				t.Errorf("ParseAndValidateDate(%q) error = %v, expected err %v", tt.input, err, tt.err)
				return
			}
			if !tt.err && !got.Equal(tt.expected) {
				t.Errorf("ParseAndValidateDate(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseAndValidateTimestamp(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	tests := []struct {
		name     string
		input    string
		expected time.Time
		err      bool
	}{
		{
			name:     "today",
			input:    "today",
			expected: today,
			err:      false,
		},
		{
			name:     "tommorow",
			input:    "tommorow", // Testing with the typo as it exists in the source
			expected: today.AddDate(0, 0, 1),
			err:      false,
		},
		{
			name:     "next week",
			input:    "next week",
			expected: today.AddDate(0, 0, 7),
			err:      false,
		},
		{
			name:     "next month",
			input:    "next month",
			expected: today.AddDate(0, 1, 0),
			err:      false,
		},
		{
			name:     "next monday",
			input:    "next monday",
			expected: NearestWeekday(today, time.Monday),
			err:      false,
		},
		{
			name:     "next saturday",
			input:    "next saturday",
			expected: NearestWeekday(today, time.Saturday),
			err:      false,
		},
		{
			name:     "in 3 days",
			input:    "in 3 days",
			expected: today.AddDate(0, 0, 3),
			err:      false,
		},
		{
			name:     "in 2 weeks",
			input:    "in 2 weeks",
			expected: today.AddDate(0, 0, 14),
			err:      false,
		},
		{
			name:     "in 1 month",
			input:    "in 1 month",
			expected: today.AddDate(0, 1, 0),
			err:      false,
		},
		{
			name:     "in 5 years",
			input:    "in 5 years",
			expected: today.AddDate(5, 0, 0),
			err:      false,
		},
		{
			name:     "explicit date with time (parsed by `layout`)",
			input:    "2025-05-10 10:00",
			expected: time.Date(2025, 5, 10, 10, 0, 0, 0, time.Local),
			err:      false,
		},
		// --- Test cases highlighting current limitations/bugs in ParseAndValidateDate ---
		{
			name:     "explicit date only",
			input:    "2025-05-10",
			expected: time.Date(2025, 5, 10, 0, 0, 0, 0, now.Location()),
			err:      false,
		},
		{
			name:     "invalid relative format: next only",
			input:    "next",
			expected: time.Time{},
			err:      true,
		},
		{
			name:     "invalid relative format: in only",
			input:    "in",
			expected: time.Time{},
			err:      true,
		},
		{
			name:     "invalid date string",
			input:    "some random string",
			expected: time.Time{},
			err:      true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: time.Time{},
			err:      true,
		},
		{
			name:     "in 3 day",
			input:    "in 3 day",
			expected: today.AddDate(0, 0, 3),
			err:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAndValidateTimestamp(tt.input)
			if (err != nil) != tt.err {
				t.Errorf("ParseAndValidateTimestamp(%q) error = %v, err %v", tt.input, err, tt.err)
				return
			}
			if !tt.err && !got.Equal(tt.expected) {
				t.Errorf("ParseAndValidateTimestamp(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
