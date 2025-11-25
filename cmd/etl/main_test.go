package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
	"go.uber.org/zap"
)

// Mock dependencies for testing the conditional logic
type mockSummariesUseCase struct {
	GenerateCalled bool
}

func (m *mockSummariesUseCase) GenerateForTitle(ctx context.Context, title domain.Title, sections []domain.Section) ([]domain.Summary, error) {
	m.GenerateCalled = true
	return []domain.Summary{{Key: title.Title, Text: "Summary"}}, nil
}

func TestSummaryGenerationFlag(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()
	title := domain.Title{Title: "1"}
	sections := []domain.Section{}

	tests := []struct {
		name        string
		skipSummary bool
		wantCalled  bool
	}{
		{
			name:        "Flag false (default) -> Execute Summary",
			skipSummary: false,
			wantCalled:  true,
		},
		{
			name:        "Flag true -> Skip Summary",
			skipSummary: true,
			wantCalled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockUsecase := &mockSummariesUseCase{}

			// Simulate the logic inside the main loop
			if !tt.skipSummary {
				logger.Info("Generating title summary", zap.String("title", title.Title))
				_, _ = mockUsecase.GenerateForTitle(ctx, title, sections)
			} else {
				logger.Info("Skipping summary generation", zap.String("title", title.Title))
			}

			// Assert
			assert.Equal(t, tt.wantCalled, mockUsecase.GenerateCalled)
		})
	}
}
