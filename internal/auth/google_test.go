package auth

import "testing"

func TestGoogleIdentityFromClaimsRequiresVerifiedEmail(t *testing.T) {
	tests := []struct {
		name     string
		claims   map[string]interface{}
		verified bool
	}{
		{
			name: "verified email",
			claims: map[string]interface{}{
				"email":          "member@example.com",
				"sub":            "google-subject",
				"email_verified": true,
			},
			verified: true,
		},
		{
			name: "unverified email",
			claims: map[string]interface{}{
				"email": "member@example.com",
				"sub":   "google-subject",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, verified := googleIdentityFromClaims(test.claims)
			if verified != test.verified {
				t.Fatalf("verified = %t, want %t", verified, test.verified)
			}
		})
	}
}

func TestGoogleIDMatchesStoredIdentity(t *testing.T) {
	storedID := "google-subject"
	tests := []struct {
		name     string
		stored   *string
		received string
		matches  bool
	}{
		{name: "same subject", stored: &storedID, received: storedID, matches: true},
		{name: "different subject", stored: &storedID, received: "other-subject"},
		{name: "missing stored subject", received: storedID},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if matches := googleIDMatches(test.stored, test.received); matches != test.matches {
				t.Fatalf("matches = %t, want %t", matches, test.matches)
			}
		})
	}
}
