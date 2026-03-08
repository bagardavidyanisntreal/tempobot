package storage

import "database/sql"

type ParticipantRepo struct {
	db *sql.DB
}

func NewParticipantRepo(db *sql.DB) *ParticipantRepo {
	return &ParticipantRepo{db: db}
}

func (r *ParticipantRepo) Upsert(eventID, userID int64, status string) error {
	query := `
		INSERT INTO participants (event_id, user_id, status)
		VALUES ($1,$2,$3)
		ON CONFLICT (event_id,user_id)
		DO UPDATE SET
		status = EXCLUDED.status,
		updated_at = now()
	`

	_, err := r.db.Exec(query, eventID, userID, status)

	return err
}

func (r *ParticipantRepo) CountByStatus(eventID int64) (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT status, count(*)
		FROM participants
		WHERE event_id = $1
		GROUP BY status
	`, eventID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	stats := map[string]int{
		"going": 0,
		"maybe": 0,
		"no":    0,
	}

	for rows.Next() {

		var status string
		var count int

		if err = rows.Scan(&status, &count); err != nil {
			return nil, err
		}

		stats[status] = count
	}

	return stats, nil
}
