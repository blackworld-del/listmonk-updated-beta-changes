package models

import null "gopkg.in/volatiletech/null.v6"

type SMTPProfiles []SMTPProfile

type SMTPProfile struct {
	Base

	UUID      string `db:"uuid" json:"uuid"`
	Name      string `db:"name" json:"name"`
	Host      string `db:"host" json:"host"`
	Port      int    `db:"port" json:"port"`
	Username  string `db:"username" json:"username"`
	Password  string `db:"password" json:"password,omitempty"`
	Encryption string `db:"encryption" json:"encryption"`
	FromEmail string `db:"from_email" json:"from_email"`
	FromName  string `db:"from_name" json:"from_name"`
	ReplyTo   string `db:"reply_to" json:"reply_to"`
	Enabled   bool   `db:"enabled" json:"enabled"`

	CampaignCount int `db:"campaign_count" json:"campaign_count"`
}

type SMTPProfileOptIn struct {
	FromEmail null.String `json:"from_email"`
	FromName  string      `json:"from_name"`
	ReplyTo   string      `json:"reply_to"`
}
