package usecase

import (
	"testing"
)

func TestCountDefs(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{"No defs", "This is a normal text.", 0},
		{"Explicit definition", "definitions. as used in this part:", 1},
		{"Means definition", "the term apple means a fruit.", 1},
		{"Combined", "definitions. term means something.", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countDefs(tt.text)
			if got != tt.want {
				t.Errorf("countDefs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeText(t *testing.T) {
	input := "Hello,   World! This is  a test."
	want := "hello world this is a test"
	got := normalizeText(input)
	if got != want {
		t.Errorf("normalizeText() = %v, want %v", got, want)
	}
}

func TestCountXrefs(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"see § 123.45", 1},
		{"refer to 40 CFR 123.45", 1},
		{"§ 1.1 and § 2.2", 2},
		{"no refs here", 0},
	}

	for _, tt := range tests {
		got := countXrefs(tt.input)
		if got != tt.expected {
			t.Errorf("countXrefs(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestCountModals(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"you shall do this", 1},
		{"you must do that", 1},
		{"you may not do this", 1},
		{"you must not do that", 1},
		{"shall we? must we?", 2},
		{"no modals here", 0},
	}

	for _, tt := range tests {
		got := countModals(tt.input)
		if got != tt.expected {
			t.Errorf("countModals(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}
