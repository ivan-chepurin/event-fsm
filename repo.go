package event_fsm

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type dbStore struct {
	db *sqlx.DB

	stateRepo
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

func (s *stateRepo) getLastLogByTargetID(ctx context.Context, targetID string) (Log, error) {
	const query = `
		SELECT 
			id,	
			target_id,
			event_id,
			current_state,
			current_result_status,
			created_at,
			updated_at 
		FROM fsm_target_logs 
		WHERE target_id = $1 
		ORDER BY created_at DESC, id DESC  -- Добавили id для детерминированности
		LIMIT 1
	`

	var log logDto
	err := s.store.db.QueryRowContext(ctx, query, targetID).Scan(&log) // Используем QueryRowContext и StructScan
	if err != nil {
		return Log{}, fmt.Errorf("failed to get last log: %w", err)
	}

	return log.toLog(), nil
}

func (s *stateRepo) saveLog(ctx context.Context, log Log) (string, error) {
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
	err := s.store.db.GetContext(ctx, &id, query, dto)
	if err != nil {
		return "", err
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

// CREATE TABLE IF NOT EXISTS fsm_target_events (
//
//	id UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
//	entity_id VARCHAR NOT NULL,
//	cur_state VARCHAR NOT NULL,
//	result_status VARCHAR NOT NULL,
//	meta_info JSONB,
//	created_at TIMESTAMPTZ DEFAULT now(),
//	updated_at TIMESTAMPTZ DEFAULT now()
//
// );
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
						target_id,
						last_result_status,
						meta_info,
						created_at,
						updated_at
					) VALUES (
						:entity_id,
						:last_result_status,
						:meta_info,
						now(),
						now()
					) RETURNING id`

	dto := eventToDTO(event)
	var id string
	err := s.store.db.GetContext(ctx, &id, query, dto)
	if err != nil {
		return "", err
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

// sqlClient - common interface for *sqlx.DB and *sqlx.TX
// https://gist.github.com/hielfx/4469d35127d085fc3501d483e34d4bad
//
//nolint:interfacebloat
type sqlClient interface {
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
	DriverName() string
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	MustExec(query string, args ...interface{}) sql.Result
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	MustExecContext(ctx context.Context, query string, args ...interface{}) sql.Result
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	Preparex(query string) (*sqlx.Stmt, error)
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	Rebind(query string) string
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
