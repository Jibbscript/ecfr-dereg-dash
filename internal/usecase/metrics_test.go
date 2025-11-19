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
