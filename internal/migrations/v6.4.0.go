package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V6_4_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS smtp_logs (
			id               BIGSERIAL PRIMARY KEY,
			smtp_profile_id  INTEGER NOT NULL REFERENCES smtp_profiles(id) ON DELETE CASCADE ON UPDATE CASCADE,
			campaign_id      INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE ON UPDATE CASCADE,
			sent             INT NOT NULL DEFAULT 0,
			failed           INT NOT NULL DEFAULT 0,
			start_time       TIMESTAMP WITH TIME ZONE,
			end_time         TIMESTAMP WITH TIME ZONE,
			created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_smtp_logs_profile ON smtp_logs(smtp_profile_id)
	`); err != nil {
		return err
	}
	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_smtp_logs_campaign ON smtp_logs(campaign_id)
	`); err != nil {
		return err
	}
	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_smtp_logs_created ON smtp_logs(created_at)
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS smtp_activity_log (
			id               BIGSERIAL PRIMARY KEY,
			smtp_profile_id  INTEGER NOT NULL REFERENCES smtp_profiles(id) ON DELETE CASCADE ON UPDATE CASCADE,
			event_type       TEXT NOT NULL DEFAULT '',
			message          TEXT NOT NULL DEFAULT '',
			created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_smtp_activity_profile ON smtp_activity_log(smtp_profile_id)
	`); err != nil {
		return err
	}
	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_smtp_activity_type ON smtp_activity_log(event_type)
	`); err != nil {
		return err
	}
	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_smtp_activity_created ON smtp_activity_log(created_at)
	`); err != nil {
		return err
	}

	return nil
}
