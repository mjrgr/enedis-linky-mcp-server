package service

import (
	"testing"
)

func TestValidateDateRange(t *testing.T) {
	tests := []struct {
		name    string
		start   string
		end     string
		wantErr bool
	}{
		{
			name:    "valid range",
			start:   "2025-01-01",
			end:     "2025-01-31",
			wantErr: false,
		},
		{
			name:    "same day",
			start:   "2025-06-15",
			end:     "2025-06-15",
			wantErr: false,
		},
		{
			name:    "invalid start format",
			start:   "01-01-2025",
			end:     "2025-01-31",
			wantErr: true,
		},
		{
			name:    "invalid end format",
			start:   "2025-01-01",
			end:     "not-a-date",
			wantErr: true,
		},
		{
			name:    "start after end",
			start:   "2025-02-01",
			end:     "2025-01-01",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDateRange(tt.start, tt.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDateRange(%q, %q) error = %v, wantErr %v",
					tt.start, tt.end, err, tt.wantErr)
			}
		})
	}
}

func TestRoundTo(t *testing.T) {
	cases := []struct {
		input    float64
		decimals int
		want     float64
	}{
		{1234.5678, 2, 1234.57},
		{1000.0, 3, 1000.0},
		{0.0, 2, 0.0},
		{-5.555, 2, -5.56},
	}

	for _, c := range cases {
		got := roundTo(c.input, c.decimals)
		if got != c.want {
			t.Errorf("roundTo(%v, %d) = %v, want %v", c.input, c.decimals, got, c.want)
		}
	}
}
