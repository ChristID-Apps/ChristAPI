package sites

import "testing"

func TestValidateCoordinatePair(t *testing.T) {
	validLat, validLon := 47.6, -122.3
	invalidLat := 91.0
	tests := []struct {
		name                   string
		lat, lon               *float64
		latPresent, lonPresent bool
		wantErr                bool
	}{
		{name: "both omitted"},
		{name: "valid pair", lat: &validLat, lon: &validLon, latPresent: true, lonPresent: true},
		{name: "one coordinate omitted", lat: &validLat, latPresent: true, wantErr: true},
		{name: "invalid latitude", lat: &invalidLat, lon: &validLon, latPresent: true, lonPresent: true, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateCoordinatePair(test.lat, test.lon, test.latPresent, test.lonPresent)
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr %t", err, test.wantErr)
			}
		})
	}
}
