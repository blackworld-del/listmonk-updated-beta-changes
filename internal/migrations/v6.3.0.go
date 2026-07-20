package migrations

import (
	"encoding/json"
	"log"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V6_3_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	// Create smtp_profiles table if it doesn't exist.
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS smtp_profiles (
			id              SERIAL PRIMARY KEY,
			uuid uuid       NOT NULL UNIQUE,
			name            TEXT NOT NULL UNIQUE,
			host            TEXT NOT NULL DEFAULT '',
			port            INT NOT NULL DEFAULT 587,
			username        TEXT NOT NULL DEFAULT '',
			password        TEXT NOT NULL DEFAULT '',
			encryption      TEXT NOT NULL DEFAULT 'starttls',
			from_email      TEXT NOT NULL DEFAULT '',
			from_name       TEXT NOT NULL DEFAULT '',
			reply_to        TEXT NOT NULL DEFAULT '',
			enabled         BOOLEAN NOT NULL DEFAULT true,
			created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`); err != nil {
		return err
	}

	// Add smtp_profile_id to campaigns if it doesn't exist.
	if _, err := db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'campaigns' AND column_name = 'smtp_profile_id'
			) THEN
				ALTER TABLE campaigns ADD COLUMN smtp_profile_id INTEGER REFERENCES smtp_profiles(id) ON DELETE SET NULL;
			END IF;
		END $$;
	`); err != nil {
		return err
	}

	// Migrate existing SMTP settings to smtp_profiles table.
	// Read the existing SMTP config from the settings table.
	var smtpJSON []byte
	if err := db.Get(&smtpJSON, `SELECT value FROM settings WHERE key = 'smtp'`); err != nil {
		// If there's no smtp setting, this is a fresh install. Nothing to migrate.
		return nil
	}

	var smtpServers []struct {
		UUID          string `json:"uuid"`
		Enabled       bool   `json:"enabled"`
		Host          string `json:"host"`
		Port          int    `json:"port"`
		AuthProtocol  string `json:"auth_protocol"`
		Username      string `json:"username"`
		Password      string `json:"password"`
		HelloHostname string `json:"hello_hostname"`
		TLSType       string `json:"tls_type"`
		TLSSkipVerify bool   `json:"tls_skip_verify"`
		FromAddresses []string `json:"from_addresses"`
		MaxConns      int    `json:"max_conns"`
		MaxMsgRetries int    `json:"max_msg_retries"`
		MsgRetryDelay string `json:"msg_retry_delay"`
		IdleTimeout   string `json:"idle_timeout"`
		WaitTimeout   string `json:"wait_timeout"`
		EmailHeaders  []map[string]string `json:"email_headers"`
		Name          string `json:"name"`
	}

	if err := json.Unmarshal(smtpJSON, &smtpServers); err != nil {
		return err
	}

	if len(smtpServers) == 0 {
		return nil
	}

	// Check if we already have smtp_profiles (idempotent).
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM smtp_profiles`); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	for _, s := range smtpServers {
		if !s.Enabled {
			continue
		}

		// Map TLS type to encryption.
		encryption := "starttls"
		switch s.TLSType {
		case "TLS":
			encryption = "ssl_tls"
		case "STARTTLS":
			encryption = "starttls"
		case "", "none":
			encryption = "none"
		}

		// Use server name or generate one.
		name := s.Name
		if name == "" {
			name = s.Host
			if name == "" {
				name = "Default SMTP"
			}
		}

		// Determine from_email from from_addresses or host.
		fromEmail := ""
		if len(s.FromAddresses) > 0 {
			fromEmail = s.FromAddresses[0]
		}

		uu, err := uuid.NewV4()
		if err != nil {
			return err
		}

		if _, err := db.Exec(`
			INSERT INTO smtp_profiles (uuid, name, host, port, username, password, encryption, from_email, enabled)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (name) DO NOTHING
		`, uu.String(), name, s.Host, s.Port, s.Username, s.Password, encryption, fromEmail, true); err != nil {
			return err
		}
	}

	// Update existing campaigns to point to the default SMTP profile.
	// Find the first enabled SMTP profile and assign it to campaigns that have
	// no smtp_profile_id and use the "email" messenger.
	if _, err := db.Exec(`
		UPDATE campaigns
		SET smtp_profile_id = (
			SELECT id FROM smtp_profiles ORDER BY id ASC LIMIT 1
		)
		WHERE smtp_profile_id IS NULL
		AND messenger = 'email'
	`); err != nil {
		return err
	}

	return nil
}
