package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type ParticipantRepo struct {
	db *sql.DB
}

func NewParticipantRepo(db *sql.DB) *ParticipantRepo {
	return &ParticipantRepo{db: db}
}

func (r *ParticipantRepo) Upsert(ctx context.Context, eventID, userID int64, status string) error {
	query := `
		INSERT INTO participants (event_id, user_id, status)
		VALUES ($1,$2,$3)
		ON CONFLICT (event_id,user_id)
		DO UPDATE SET
		status = EXCLUDED.status,
		updated_at = now()
	`

	_, err := r.db.ExecContext(ctx, query, eventID, userID, status)
	if err != nil {
		return fmt.Errorf("upsert participant: %w", err)
	}

	return nil
}

func (r *ParticipantRepo) CountByStatus(ctx context.Context, eventID int64) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT status, count(*)
		FROM participants
		WHERE event_id = $1
		GROUP BY status
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("countByStatus participants: %w", err)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("countByStatus participants: %w", rows.Err())
	}

	defer func() { _ = rows.Close() }()

	stats := map[string]int{
		"going": 0,
		"maybe": 0,
		"no":    0,
	}

	for rows.Next() {
		var (
			status string
			count  int
		)

		if err = rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("countByStatus participants: %w", err)
		}

		stats[status] = count
	}

	return stats, nil
}
