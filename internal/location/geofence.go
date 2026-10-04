package location

import (
	"errors"
	"math"
	"sort"
	"strconv"
)

const (
	GeofenceRadiusMeters = 500.0
	earthRadiusMeters    = 6_371_000.0
)

var (
	ErrInvalidPosition     = errors.New("invalid device location or accuracy")
	ErrNoNearbyTarget      = errors.New("no attendance location is within range")
	ErrTargetNotConfigured = errors.New("attendance location is not configured")
)

type Position struct {
	Latitude  float64
	Longitude float64
	Accuracy  float64
}

type Target struct {
	Type      string
	ID        int64
	Name      string
	Latitude  float64
	Longitude float64
}

type Match struct {
	Target   Target
	Distance float64
}

func ValidatePosition(position Position) error {
	if !validCoordinate(position.Latitude, position.Longitude) ||
		math.IsNaN(position.Accuracy) || math.IsInf(position.Accuracy, 0) ||
		position.Accuracy < 0 {
		return ErrInvalidPosition
	}
	return nil
}

func NearestTarget(position Position, targets []Target) (*Match, error) {
	if err := ValidatePosition(position); err != nil {
		return nil, err
	}
	matches := make([]Match, 0, len(targets))
	for _, target := range targets {
		if (target.ID < 1 && target.Type != "configured") || !validCoordinate(target.Latitude, target.Longitude) {
			continue
		}
		distance := DistanceMeters(position.Latitude, position.Longitude, target.Latitude, target.Longitude)
		if distance <= GeofenceRadiusMeters+1e-6 {
			matches = append(matches, Match{Target: target, Distance: math.Min(distance, GeofenceRadiusMeters)})
		}
	}
	if len(matches) == 0 {
		return nil, ErrNoNearbyTarget
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Distance == matches[j].Distance {
			if matches[i].Target.Type == matches[j].Target.Type {
				return matches[i].Target.ID < matches[j].Target.ID
			}
			return matches[i].Target.Type < matches[j].Target.Type
		}
		return matches[i].Distance < matches[j].Distance
	})
	return &matches[0], nil
}

func ConfiguredAttendanceTarget(latitude, longitude string) (Target, error) {
	if latitude == "" || longitude == "" {
		return Target{}, ErrTargetNotConfigured
	}
	lat, err := strconv.ParseFloat(latitude, 64)
	if err != nil {
		return Target{}, ErrTargetNotConfigured
	}
	lon, err := strconv.ParseFloat(longitude, 64)
	if err != nil || !validCoordinate(lat, lon) {
		return Target{}, ErrTargetNotConfigured
	}
	return Target{Type: "configured", Name: "Gereja", Latitude: lat, Longitude: lon}, nil
}

func DistanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	toRadians := math.Pi / 180
	phi1 := lat1 * toRadians
	phi2 := lat2 * toRadians
	deltaPhi := (lat2 - lat1) * toRadians
	deltaLambda := (lon2 - lon1) * toRadians
	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	return 2 * earthRadiusMeters * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func validCoordinate(latitude, longitude float64) bool {
	return !math.IsNaN(latitude) && !math.IsInf(latitude, 0) && latitude >= -90 && latitude <= 90 &&
		!math.IsNaN(longitude) && !math.IsInf(longitude, 0) && longitude >= -180 && longitude <= 180
}
