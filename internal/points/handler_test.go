package points

import "testing"

func TestParseEarnTargetUserID(t *testing.T) {
	validID := int64(27)
	zeroID := int64(0)
	tests := []struct {
		name    string
		userID  *int64
		wantID  int64
		wantErr bool
	}{
		{name: "missing user id", wantErr: true},
		{name: "non-positive user id", userID: &zeroID, wantErr: true},
		{name: "valid user id", userID: &validID, wantID: validID},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseEarnTargetUserID(test.userID)
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr = %t", err, test.wantErr)
			}
			if got != test.wantID {
				t.Fatalf("user id = %d, want %d", got, test.wantID)
			}
		})
	}
}
