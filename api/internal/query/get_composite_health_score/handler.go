// Package get_composite_health_score menangani query untuk mengambil composite health score.
package get_composite_health_score

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_composite_health_score"

// Query berisi parameter untuk mengambil composite health score.
type Query struct{}

func (q Query) QueryName() string { return queryName }

// CategoryScore mewakili skor untuk satu kategori KPI.
type CategoryScore struct {
	Category string  `json:"category"`
	Score    float64 `json:"score"`
	Count    int     `json:"count"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	OverallScore    float64         `json:"overall_score"`
	CategoryScores  []*CategoryScore `json:"category_scores"`
	TotalIndicators int             `json:"total_indicators"`
}

// Handler menangani Query get_composite_health_score.
type Handler struct {
	repo kpi.KPIDefinitionReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo kpi.KPIDefinitionReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil composite health score dari semua KPI aktif.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	isActive := true
	entities, total, err := h.repo.List(ctx, s, kpi.KPIDefinitionFilter{IsActive: &isActive}, 500, 0, "category", "asc")
	if err != nil {
		return nil, fmt.Errorf("list kpis for health score: %w", err)
	}

	categoryMap := make(map[string][]float64)
	for _, e := range entities {
		// Skor sederhana berdasarkan threshold: rata-rata target vs thresholds.
		score := calculateScore(e)
		categoryMap[string(e.Category)] = append(categoryMap[string(e.Category)], score)
	}

	categoryScores := make([]*CategoryScore, 0, len(categoryMap))
	var totalScore float64
	var totalCount int

	for cat, scores := range categoryMap {
		avg := average(scores)
		categoryScores = append(categoryScores, &CategoryScore{
			Category: cat,
			Score:    avg,
			Count:    len(scores),
		})
		totalScore += avg
		totalCount++
	}

	overall := float64(0)
	if totalCount > 0 {
		overall = totalScore / float64(totalCount)
	}

	return &Result{
		OverallScore:    overall,
		CategoryScores:  categoryScores,
		TotalIndicators: total,
	}, nil
}

func calculateScore(e *kpi.KPIDefinition) float64 {
	if e.TargetValue == 0 {
		return 100
	}
	// Simplified scoring: ratio of target to warning threshold.
	ratio := e.TargetValue / e.WarningThreshold
	if ratio > 1 {
		return 100
	}
	return ratio * 100
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
