package biblereadings

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestValidatePayloadAcceptsOnlyBoundedJSONObjects(t *testing.T) {
	tests := []struct {
		name    string
		payload json.RawMessage
		wantErr bool
	}{
		{name: "object", payload: json.RawMessage(`{"reading":"Mazmur 23","reflection":"Tuhan memelihara"}`)},
		{name: "array", payload: json.RawMessage(`[1,2]`), wantErr: true},
		{name: "null", payload: json.RawMessage(`null`), wantErr: true},
		{name: "invalid JSON", payload: json.RawMessage(`{"reading":`), wantErr: true},
		{name: "too large", payload: json.RawMessage(`{"text":"` + string(make([]byte, maxPayloadBytes)) + `"}`), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidatePayload(test.payload)
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr %t", err, test.wantErr)
			}
			if test.wantErr && !errors.Is(err, ErrInvalidPayload) {
				t.Fatalf("error = %v, want ErrInvalidPayload", err)
			}
		})
	}
}
