package jira

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJiraTimeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantTime    time.Time
		wantErr     bool
		description string
	}{
		{
			name:        "valid Jira timestamp",
			input:       `"2024-01-01T10:00:00.000+0000"`,
			wantTime:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			wantErr:     false,
			description: "Standard Jira timestamp format",
		},
		{
			name:        "valid timestamp with negative offset",
			input:       `"2024-06-15T14:30:45.123-0700"`,
			wantTime:    time.Date(2024, 6, 15, 14, 30, 45, 123000000, time.FixedZone("", -7*3600)),
			wantErr:     false,
			description: "Timestamp with negative timezone offset",
		},
		{
			name:        "valid timestamp with positive offset",
			input:       `"2024-12-25T23:59:59.999+0530"`,
			wantTime:    time.Date(2024, 12, 25, 23, 59, 59, 999000000, time.FixedZone("", 5*3600+30*60)),
			wantErr:     false,
			description: "Timestamp with positive timezone offset",
		},
		{
			name:        "empty string",
			input:       `""`,
			wantTime:    time.Time{},
			wantErr:     false,
			description: "Empty timestamp should not error",
		},
		{
			name:        "null value",
			input:       `null`,
			wantTime:    time.Time{},
			wantErr:     false,
			description: "Null should result in zero time",
		},
		{
			name:        "invalid format - missing timezone",
			input:       `"2024-01-01T10:00:00"`,
			wantErr:     true,
			description: "Timestamp without timezone should fail",
		},
		{
			name:        "invalid format - wrong date format",
			input:       `"01-01-2024 10:00:00"`,
			wantErr:     true,
			description: "Non-ISO8601 format should fail",
		},
		{
			name:        "invalid format - malformed",
			input:       `"not-a-timestamp"`,
			wantErr:     true,
			description: "Completely malformed timestamp should fail",
		},
		{
			name:        "invalid JSON",
			input:       `{invalid}`,
			wantErr:     true,
			description: "Invalid JSON should fail at unmarshal stage",
		},
		{
			name:        "number instead of string",
			input:       `1234567890`,
			wantErr:     true,
			description: "Number format should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jt JiraTime
			err := json.Unmarshal([]byte(tt.input), &jt)

			if tt.wantErr {
				if err == nil {
					t.Errorf("JiraTime.UnmarshalJSON() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("JiraTime.UnmarshalJSON() unexpected error = %v", err)
			}

			// For empty string and null cases, check that time is zero
			if tt.input == `""` || tt.input == `null` {
				if !jt.IsZero() {
					t.Errorf("JiraTime.UnmarshalJSON() for %s should result in zero time, got %v", tt.input, jt.Time)
				}
				return
			}

			if !jt.Equal(tt.wantTime) {
				t.Errorf("JiraTime.UnmarshalJSON() time = %v, want %v", jt.Time, tt.wantTime)
			}
		})
	}
}

func TestJiraTimeUnmarshalJSONInStruct(t *testing.T) {
	// Test unmarshaling JiraTime as part of a larger struct
	type TestStruct struct {
		Created JiraTime `json:"created"`
		Updated JiraTime `json:"updated"`
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, ts TestStruct)
	}{
		{
			name:    "both timestamps valid",
			input:   `{"created":"2024-01-01T10:00:00.000+0000","updated":"2024-01-15T14:30:00.000+0000"}`,
			wantErr: false,
			check: func(t *testing.T, ts TestStruct) {
				expectedCreated := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
				expectedUpdated := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
				if !ts.Created.Equal(expectedCreated) {
					t.Errorf("Created time = %v, want %v", ts.Created.Time, expectedCreated)
				}
				if !ts.Updated.Equal(expectedUpdated) {
					t.Errorf("Updated time = %v, want %v", ts.Updated.Time, expectedUpdated)
				}
			},
		},
		{
			name:    "one timestamp empty",
			input:   `{"created":"2024-01-01T10:00:00.000+0000","updated":""}`,
			wantErr: false,
			check: func(t *testing.T, ts TestStruct) {
				if ts.Created.IsZero() {
					t.Error("Created time should not be zero")
				}
				if !ts.Updated.IsZero() {
					t.Error("Updated time should be zero for empty string")
				}
			},
		},
		{
			name:    "one timestamp invalid",
			input:   `{"created":"2024-01-01T10:00:00.000+0000","updated":"invalid"}`,
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts TestStruct
			err := json.Unmarshal([]byte(tt.input), &ts)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.check != nil {
				tt.check(t, ts)
			}
		})
	}
}

func TestJiraTimeEdgeCases(t *testing.T) {
	t.Run("leap year date", func(t *testing.T) {
		input := `"2024-02-29T12:00:00.000+0000"`
		var jt JiraTime
		err := json.Unmarshal([]byte(input), &jt)
		if err != nil {
			t.Fatalf("Failed to unmarshal leap year date: %v", err)
		}
		expected := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)
		if !jt.Equal(expected) {
			t.Errorf("Leap year date = %v, want %v", jt.Time, expected)
		}
	})

	t.Run("midnight", func(t *testing.T) {
		input := `"2024-01-01T00:00:00.000+0000"`
		var jt JiraTime
		err := json.Unmarshal([]byte(input), &jt)
		if err != nil {
			t.Fatalf("Failed to unmarshal midnight: %v", err)
		}
		expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		if !jt.Equal(expected) {
			t.Errorf("Midnight time = %v, want %v", jt.Time, expected)
		}
	})

	t.Run("end of day", func(t *testing.T) {
		input := `"2024-12-31T23:59:59.999+0000"`
		var jt JiraTime
		err := json.Unmarshal([]byte(input), &jt)
		if err != nil {
			t.Fatalf("Failed to unmarshal end of day: %v", err)
		}
		expected := time.Date(2024, 12, 31, 23, 59, 59, 999000000, time.UTC)
		if !jt.Equal(expected) {
			t.Errorf("End of day time = %v, want %v", jt.Time, expected)
		}
	})
}
