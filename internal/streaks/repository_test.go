package streaks

import (
	"errors"
	"testing"
)

func TestValidateAttendanceRecord(t *testing.T) {
	if err := validateAttendanceRecord(true); err != nil {
		t.Fatalf("valid attendance returned error: %v", err)
	}
	if err := validateAttendanceRecord(false); !errors.Is(err, ErrAttendanceRequired) {
		t.Fatalf("missing attendance error = %v, want %v", err, ErrAttendanceRequired)
	}
}
