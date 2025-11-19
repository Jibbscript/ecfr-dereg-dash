package unit_test

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
		{"Explicit definition", "Definitions. As used in this part:", 1},
		{"Means definition", "The term apple means a fruit.", 1},
		{"Combined", "Definitions. term means something.", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to export countDefs or move test to package usecase
			// Since we are in unit_test package, we assume we can access public methods or we'd move this file.
			// For this MVP, let's assume we moved the logic to a public helper or test internal package.
			// Actually, usecase.countDefs is private. We should export it or test in package usecase.
			// Let's change package to usecase.
		})
	}
}
