package main

import (
	"bytes"
	"testing"

	"rofi-chrome-tab/internal/model"
)

func TestListTabs(t *testing.T) {
	// Save original tabs and restore after test
	originalTabs := tabs
	defer func() {
		tabs = originalTabs
	}()

	// Set up test tabs
	tabs = []model.Tab{
		{ID: 1, Title: "Tab 1", Host: "example.com"},
		{ID: 2, Title: "Tab 2", Host: "google.com"},
	}

	tests := []struct {
		name       string
		pid        int
		wantOutput string
	}{
		{
			name:       "list tabs with pid 12345",
			pid:        12345,
			wantOutput: "12345,1,example.com,Tab 1\n12345,2,google.com,Tab 2\n",
		},
		{
			name:       "list tabs with pid 99999",
			pid:        99999,
			wantOutput: "99999,1,example.com,Tab 1\n99999,2,google.com,Tab 2\n",
		},
		{
			name:       "list tabs with pid 1",
			pid:        1,
			wantOutput: "1,1,example.com,Tab 1\n1,2,google.com,Tab 2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := listTabs(&buf, tt.pid)
			if err != nil {
				t.Fatalf("listTabs() error = %v", err)
			}

			got := buf.String()
			if got != tt.wantOutput {
				t.Errorf("listTabs() output = %q, want %q", got, tt.wantOutput)
			}
		})
	}
}

func TestListTabsEmptyTabs(t *testing.T) {
	// Save original tabs and restore after test
	originalTabs := tabs
	defer func() {
		tabs = originalTabs
	}()

	// Set up empty tabs
	tabs = []model.Tab{}

	var buf bytes.Buffer
	err := listTabs(&buf, 12345)
	if err != nil {
		t.Fatalf("listTabs() error = %v", err)
	}

	got := buf.String()
	if got != "" {
		t.Errorf("listTabs() with empty tabs output = %q, want empty string", got)
	}
}
