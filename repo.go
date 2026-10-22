package event_fsm

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type dbStore struct {
	db *sqlx.DB
}

func newDBStore(db *sqlx.DB) *dbStore {
	s := &dbStore{
		db: db,
	}

	return s
}

type stateRepo struct {
	store *dbStore
}

func newRepo(store *dbStore) *stateRepo {
	return &stateRepo{
		store: store,
	}
}

func (s *stateRepo) createLog(ctx context.Context, log Log) (string, error) {
	const query = `INSERT INTO fsm_target_logs (
						target_id,
						event_id,
						current_state,
						created_at,
						updated_at
                  	) VALUES (
						:target_id,
					  	:event_id,
						:current_state,
						now(), 
						now()
					) RETURNING id`

	dto := logToDTO(log)
	var id string
	rows, err := s.store.db.NamedQueryContext(ctx, query, dto)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("no rows returned")
	}

	return id, nil
}

func (s *stateRepo) createFullLog(ctx context.Context, log Log) (string, error) {
	const query = `INSERT INTO fsm_target_logs (
						target_id,
						event_id,
						current_state,
						current_result_status,
						created_at,
						updated_at
				  	) VALUES (
						:target_id,
					  	:event_id,
						:current_state,
						:current_result_status,
						now(), 
						now()
					) RETURNING id`

	dto := logToDTO(log)
	var id string
	rows, err := s.store.db.NamedQueryContext(ctx, query, dto)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("no rows returned")
	}

	return id, nil
}

func (s *stateRepo) updateLog(ctx context.Context, log Log) error {
	const query = `UPDATE fsm_target_logs
					SET current_result_status = :current_result_status,
						updated_at = now()
					WHERE id = :id`

	dto := logToDTO(log)
	_, err := s.store.db.NamedExecContext(ctx, query, dto)
	if err != nil {
		return err
	}

	return nil
}

// deleteLogs deletes logs older than the specified duration,
// keeping the last 'keepCount' logs for each 'id' (source).
//
//nolint:unused // will be exposed through the public read API
func (s *stateRepo) deleteLogs(ctx context.Context, duration time.Duration, keepCount int) error {
	const query = `
		DELETE FROM fsm_target_logs
		WHERE id IN (
			SELECT id
			FROM (
				SELECT
					id,
					ROW_NUMBER() OVER (PARTITION BY target_id ORDER BY created_at DESC) as rn
				FROM fsm_target_logs
				WHERE created_at < NOW() - $1
			) as sub
			WHERE rn > $2
		);
	`

	_, err := s.store.db.ExecContext(ctx, query, duration, keepCount)
	if err != nil {
		return fmt.Errorf("failed to delete logs: %w", err)
	}

	return nil
}

//nolint:unused // will be exposed through the public read API
func (s *stateRepo) getEventByID(ctx context.Context, id string) (Event, error) {
	const query = `	SELECT
						id,
						target_id,
						last_result_status,
						meta_info,
						created_at,
						updated_at
					FROM fsm_target_events
					WHERE id = $1`

	var dto eventDto
	err := s.store.db.GetContext(ctx, &dto, query, id)
	if err != nil {
		return Event{}, err
	}

	return dto.toEvent(), nil
}

func (s *stateRepo) createEvent(ctx context.Context, event Event) (string, error) {
	const query = `INSERT INTO fsm_target_events (
					   	id,
						target_id,
						last_result_status,
						meta_info,
						created_at,
						updated_at
					) VALUES (
					    :id,
						:target_id,
						:last_result_status,
						:meta_info,
						now(),
						now()
					) RETURNING id`

	dto := eventToDTO(event)
	var id string
	rows, err := s.store.db.NamedQueryContext(ctx, query, dto)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("no rows returned")
	}

	return id, nil
}

func (s *stateRepo) updateEvent(ctx context.Context, event Event) error {
	const query = `UPDATE fsm_target_events
					SET last_result_status = :last_result_status,
						updated_at = now()
					WHERE id = :id`

	dto := eventToDTO(event)
	_, err := s.store.db.NamedExecContext(ctx, query, dto)
	if err != nil {
		return err
	}

	return nil
}
