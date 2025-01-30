package repositories

import (
	"context"
	"database/sql"
	"time"
)

type statsRepo struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) StatsRepository {
	return &statsRepo{db: db}
}

func (s *statsRepo) CountRevNonRevImages(ctx context.Context) (int, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	counts := map[bool]int{true: 0, false: 0}

	rows, err := s.db.QueryContext(ctx, `
		SELECT reviewed, count(1) AS count
		FROM uploaded_images
		GROUP BY reviewed
	`)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var reviewed bool
		var count int
		if err := rows.Scan(&reviewed, &count); err != nil {
			return 0, 0, err
		}
		counts[reviewed] = count
	}

	if err := rows.Err(); err != nil {
		return 0, 0, err
	}

	return counts[true], counts[false], nil
}

func (s *statsRepo) AverageConfidence(ctx context.Context) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var avgConfidence float64
	err := s.db.QueryRowContext(ctx, `
		SELECT AVG(confidence) 
		FROM uploaded_images
	`).Scan(&avgConfidence)
	if err != nil {
		return 0, err
	}

	return avgConfidence, nil
}

func (s *statsRepo) LabelDistribution(ctx context.Context) (map[string]int, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	labelCounts := make(map[string]int)

	rows, err := s.db.QueryContext(ctx, `
		SELECT new_label, COUNT(1) 
		FROM uploaded_images 
		WHERE reviewed = true 
		GROUP BY label
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var label string
		var count int
		if err := rows.Scan(&label, &count); err != nil {
			return nil, err
		}
		labelCounts[label] = count
	}

	return labelCounts, nil
}

func (s *statsRepo) LabelingEfficiency(ctx context.Context) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var labeledCount, totalCount int

	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1) 
		FROM uploaded_images 
		WHERE reviewed = true AND new_label = label
	`).Scan(&labeledCount)
	if err != nil {
		return 0, err
	}

	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(1) 
		FROM uploaded_images
	`).Scan(&totalCount)
	if err != nil {
		return 0, err
	}

	if totalCount == 0 {
		return 0, nil
	}

	efficiency := float64(labeledCount) / float64(totalCount) * 100.0
	return efficiency, nil
}
