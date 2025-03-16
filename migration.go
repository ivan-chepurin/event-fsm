package event_fsm

import (
	"errors"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

const (
	migrationPath      = "fsm_migrations/"
	migrationTableName = "fsm_schema_migrations"
)

type migrationData struct {
	Version string // 0001, 0002, etc
	Name    string
	Type    string // up or down
	Data    string
}

// Migrate runs the migration files in the migrationPath folder.
// It takes the database connection
// db - sqlx database connection
func Migrate(db *sqlx.DB) error {
	if err := createMigrationDirWithFiles(migrationPath, migrations); err != nil {
		return err
	}

	migrationConfigs := &postgres.Config{
		MigrationsTable: migrationTableName,
	}
	if err := migrateUp(db, migrationPath, migrationConfigs); err != nil {
		return err
	}

	if err := removeMigrationsFiles(migrationPath); err != nil {
		return err
	}

	return nil
}

// createMigrationDirWithFiles creates the migration dir and files with the given data
func createMigrationDirWithFiles(migrationPath string, migrations []migrationData) error {
	if err := os.Mkdir(migrationPath, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	for _, m := range migrations {
		fileName := migrationPath + m.Version + "_" + m.Name + "." + m.Type + ".sql"

		file, err := os.Create(fileName)
		if err != nil {
			return err
		}

		if _, err = file.WriteString(m.Data); err != nil {
			return err
		}

		if err = file.Close(); err != nil {
			return err
		}
	}

	return nil
}

// migrateUp runs the migration files in the given path
func migrateUp(db *sqlx.DB, migrationPath string, pgConfig *postgres.Config) error {
	dr, err := postgres.WithInstance(db.DB, pgConfig)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationPath, "postgres", dr)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// removeMigrationsFiles removes the migration files including directory
func removeMigrationsFiles(migrationPath string) error {
	if err := os.RemoveAll(migrationPath); err != nil {
		return err
	}

	return nil
}

//	type eventDto struct {
//		TargetID           string          `db:"id" json:"id"`
//		TargetID     string          `db:"target_id" json:"target_id"`
//		CurrentState StateName       `db:"cur_state" json:"current_state"`
//		ResultStatus ResultStatus    `db:"result_status" json:"result_status"`
//		MetaInfo     json.RawMessage `db:"meta_info" json:"meta_info"`
//		CreatedAt    time.Time       `db:"created_at" json:"created_at"`
//		UpdatedAt    time.Time       `db:"updated_at" json:"updated_at"`
//	}
//
//	type logDto struct {
//		TargetID            string       `db:"id" json:"id"`
//		EventID       string       `db:"event_id" json:"event_id"`
//		CurrentState  StateName    `db:"current_state" json:"current_state"`
//		CurrentResult ResultStatus `db:"current_result_status" json:"current_result_status"`
//		RetryCount    int          `db:"retry_count" json:"retry_count"`
//		LastRetryAt   time.Time    `db:"last_retry_at" json:"last_retry_at"`
//		CreatedAt     time.Time    `db:"created_at" json:"created_at"`
//		UpdatedAt     time.Time    `db:"updated_at" json:"updated_at"`
//	}
var migrations = []migrationData{
	{
		Version: "0001",
		Name:    "create_tables",
		Type:    "up",
		Data: `
			BEGIN;
			
			CREATE TABLE IF NOT EXISTS fsm_target_events (
				id UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
				target_id VARCHAR NOT NULL,
				last_result_status VARCHAR NOT NULL,
				meta_info JSONB,
				created_at TIMESTAMPTZ DEFAULT now(),
				updated_at TIMESTAMPTZ DEFAULT now()
			);
			
			CREATE INDEX IF NOT EXISTS fsm_target_events_idx ON fsm_target_events (id);
			CREATE INDEX IF NOT EXISTS fsm_target_events_target_id_idx ON fsm_target_events (target_id);
			
			CREATE TABLE IF NOT EXISTS fsm_target_logs (
				id UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
				target_id VARCHAR NOT NULL,
				event_id UUID NOT NULL,
				current_state VARCHAR NOT NULL,
				current_result_status VARCHAR,
				created_at TIMESTAMPTZ DEFAULT now(),
				updated_at TIMESTAMPTZ DEFAULT now()
			);
			
			CREATE INDEX IF NOT EXISTS fsm_target_logs_target_id_created_at_idx ON fsm_target_logs (target_id, created_at DESC);
			
			COMMIT;
		`,
	},
	{
		Version: "0001",
		Name:    "create_tables",
		Type:    "down",
		Data: `
			BEGIN;

			DROP TABLE IF EXISTS fsm_target_events;
			DROP TABLE IF EXISTS fsm_target_logs;
			
			COMMIT;
		`,
	},
}
