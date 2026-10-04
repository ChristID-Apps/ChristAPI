package location

import (
	"errors"
	"math"
	"testing"
)

func TestNearestTargetSelectsClosestInsideRadius(t *testing.T) {
	position := Position{Latitude: 0, Longitude: 0, Accuracy: 5}
	targets := []Target{
		{Type: "site", ID: 4, Name: "farther", Latitude: 0, Longitude: 0.002},
		{Type: "activity", ID: 7, Name: "nearer", Latitude: 0, Longitude: 0.001},
	}

	match, err := NearestTarget(position, targets)
	if err != nil {
		t.Fatal(err)
	}
	if match.Target.Name != "nearer" {
		t.Fatalf("matched target = %q, want nearer", match.Target.Name)
	}
}

func TestNearestTargetRejectsOutsideRadius(t *testing.T) {
	_, err := NearestTarget(Position{Latitude: 0, Longitude: 0, Accuracy: 1}, []Target{
		{Type: "site", ID: 1, Latitude: 0, Longitude: 0.01},
	})
	if !errors.Is(err, ErrNoNearbyTarget) {
		t.Fatalf("error = %v, want ErrNoNearbyTarget", err)
	}
}

func TestValidatePositionRejectsInvalidCoordinatesAndAccuracy(t *testing.T) {
	tests := []Position{
		{Latitude: 91, Longitude: 0, Accuracy: 1},
		{Latitude: 0, Longitude: 181, Accuracy: 1},
		{Latitude: 0, Longitude: 0, Accuracy: -1},
		{Latitude: 0, Longitude: 0, Accuracy: math.NaN()},
	}
	for _, position := range tests {
		if err := ValidatePosition(position); !errors.Is(err, ErrInvalidPosition) {
			t.Errorf("ValidatePosition(%+v) = %v, want ErrInvalidPosition", position, err)
		}
	}
}

func TestValidatePositionAllowsStandardLocationAccuracy(t *testing.T) {
	if err := ValidatePosition(Position{Latitude: -6.2375, Longitude: 106.696, Accuracy: 25000}); err != nil {
		t.Fatalf("standard location accuracy should be allowed: %v", err)
	}
}

func TestNearestTargetAcceptsExactlyAtRadius(t *testing.T) {
	longitude := GeofenceRadiusMeters / earthRadiusMeters * 180 / math.Pi
	_, err := NearestTarget(Position{Latitude: 0, Longitude: 0, Accuracy: 100}, []Target{
		{Type: "site", ID: 1, Latitude: 0, Longitude: longitude},
	})
	if err != nil {
		t.Fatalf("target at radius boundary rejected: %v", err)
	}
}

func TestConfiguredAttendanceTargetRequiresValidCoordinatePair(t *testing.T) {
	tests := []struct {
		name      string
		latitude  string
		longitude string
		wantErr   bool
	}{
		{name: "valid coordinates", latitude: "-6.2", longitude: "106.8"},
		{name: "missing longitude", latitude: "-6.2", wantErr: true},
		{name: "invalid number", latitude: "north", longitude: "106.8", wantErr: true},
		{name: "out of bounds", latitude: "91", longitude: "0", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			target, err := ConfiguredAttendanceTarget(test.latitude, test.longitude)
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr %t", err, test.wantErr)
			}
			if !test.wantErr && (target.Type != "configured" || target.Name != "Gereja" || target.ID != 0) {
				t.Fatalf("target = %+v, want configured church target", target)
			}
		})
	}
}

func TestConfiguredChurchTargetCanMatchDevicePosition(t *testing.T) {
	target, err := ConfiguredAttendanceTarget("-6.2", "106.8")
	if err != nil {
		t.Fatal(err)
	}
	match, err := NearestTarget(Position{Latitude: -6.2, Longitude: 106.8, Accuracy: 10}, []Target{target})
	if err != nil {
		t.Fatal(err)
	}
	if match.Target.Type != "configured" || match.Distance != 0 {
		t.Fatalf("match = %+v, want configured church at zero distance", match)
	}
}
