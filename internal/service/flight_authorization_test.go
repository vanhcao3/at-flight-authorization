package service

import (
	"testing"
	"time"

	"172.21.5.249/airtrans/at-flight-authorization/internal/models"
)

func mustParseTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}

func TestDurationsOverlap(t *testing.T) {
	baseStart := mustParseTime("2025-01-01T00:00:00Z")
	baseEnd := mustParseTime("2025-01-02T00:00:00Z")
	partialStart := mustParseTime("2025-01-01T12:00:00Z")
	partialEnd := mustParseTime("2025-01-02T12:00:00Z")
	touchEnd := mustParseTime("2025-01-02T12:00:00Z")
	afterStart := mustParseTime("2025-01-03T00:00:00Z")
	afterEnd := mustParseTime("2025-01-04T00:00:00Z")
	beforeStart := mustParseTime("2024-12-30T00:00:00Z")
	beforeEnd := mustParseTime("2024-12-31T00:00:00Z")

	tests := []struct {
		name string
		a    models.OperatingDuration
		b    models.OperatingDuration
		want bool
	}{
		{
			name: "identical ranges",
			a:    models.OperatingDuration{FromDay: baseStart, ToDay: baseEnd},
			b:    models.OperatingDuration{FromDay: baseStart, ToDay: baseEnd},
			want: true,
		},
		{
			name: "partial overlap",
			a:    models.OperatingDuration{FromDay: baseStart, ToDay: baseEnd},
			b:    models.OperatingDuration{FromDay: partialStart, ToDay: partialEnd},
			want: true,
		},
		{
			name: "touching edge",
			a:    models.OperatingDuration{FromDay: baseStart, ToDay: baseEnd},
			b:    models.OperatingDuration{FromDay: baseEnd, ToDay: touchEnd},
			want: true,
		},
		{
			name: "no overlap after",
			a:    models.OperatingDuration{FromDay: baseStart, ToDay: baseEnd},
			b:    models.OperatingDuration{FromDay: afterStart, ToDay: afterEnd},
			want: false,
		},
		{
			name: "no overlap before",
			a:    models.OperatingDuration{FromDay: baseStart, ToDay: baseEnd},
			b:    models.OperatingDuration{FromDay: beforeStart, ToDay: beforeEnd},
			want: false,
		},
		{
			name: "missing endpoints",
			a:    models.OperatingDuration{},
			b:    models.OperatingDuration{FromDay: baseStart, ToDay: baseEnd},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := durationsOverlap(tt.a, tt.b); got != tt.want {
				t.Fatalf("durationsOverlap() = %v, want %v", got, tt.want)
			}
			if got := durationsOverlap(tt.b, tt.a); got != tt.want {
				t.Fatalf("durationsOverlap swapped = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotificationStatusFromOverlap(t *testing.T) {
	if got := notificationStatusFromOverlap(true); got != models.FlightNotificationStatusPending {
		t.Fatalf("expected pending, got %s", got)
	}
	if got := notificationStatusFromOverlap(false); got != models.FlightNotificationStatusActivated {
		t.Fatalf("expected activated, got %s", got)
	}
}
