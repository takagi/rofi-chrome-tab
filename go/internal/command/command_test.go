package command

import (
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Command
		wantErr bool
	}{
		{"list", "list", ListCommand{}, false},
		{"select ok", "select 123", SelectCommand{TabID: 123}, false},
		{"empty", "", nil, true},
		{"unknown", "foo", nil, true},
		{"select bad arg", "select abc", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("mismatch: got=%#v want=%#v", got, tt.want)
			}
		})
	}
}
