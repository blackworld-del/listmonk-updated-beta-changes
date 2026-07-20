package models

import (
	null "gopkg.in/volatiletech/null.v6"
)

type SMTPProfileStats struct {
	TotalSent        int            `db:"total_sent" json:"total_sent"`
	TotalFailed      int            `db:"total_failed" json:"total_failed"`
	SuccessRate      float64        `db:"success_rate" json:"success_rate"`
	SentToday        int            `db:"sent_today" json:"sent_today"`
	FailedToday      int            `db:"failed_today" json:"failed_today"`
	SentWeek         int            `db:"sent_week" json:"sent_week"`
	SentMonth        int            `db:"sent_month" json:"sent_month"`
	LastSentAt       null.Time      `db:"last_sent_at" json:"last_sent_at"`
	AvgSendTimeSec   null.Float64   `db:"avg_send_time_seconds" json:"avg_send_time_seconds"`
	LastCampaignID   null.Int       `db:"last_campaign_id" json:"last_campaign_id"`
}

type SMTPDailyStat struct {
	Date        string  `db:"date" json:"date"`
	Sent        int     `db:"sent" json:"sent"`
	Failed      int     `db:"failed" json:"failed"`
	SuccessRate float64 `db:"success_rate" json:"success_rate"`
}

type SMTPRecentCampaign struct {
	ID        int        `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	StartedAt null.Time  `db:"started_at" json:"started_at"`
	EndTime   null.Time  `db:"end_time" json:"end_time"`
	Sent      int        `db:"sent" json:"sent"`
	ToSend    int        `db:"to_send" json:"to_send"`
	Status    string     `db:"status" json:"status"`
	Failed    int        `db:"failed" json:"failed"`
}

type SMTPActivity struct {
	Base
	SMTPProfileID int    `db:"smtp_profile_id" json:"smtp_profile_id"`
	EventType     string `db:"event_type" json:"event_type"`
	Message       string `db:"message" json:"message"`
}

type SMTPOverview struct {
	ID          int        `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Host        string     `db:"host" json:"host"`
	Username    string     `db:"username" json:"username"`
	Enabled     bool       `db:"enabled" json:"enabled"`
	SentToday   int        `db:"sent_today" json:"sent_today"`
	TotalSent   int        `db:"total_sent" json:"total_sent"`
	TotalFailed int        `db:"total_failed" json:"total_failed"`
	SuccessRate float64    `db:"success_rate" json:"success_rate"`
	LastSentAt  null.Time  `db:"last_sent_at" json:"last_sent_at"`
}

type SMTPDashboardStats struct {
	TotalProfiles   int `db:"total_profiles" json:"total_profiles"`
	ActiveProfiles  int `db:"active_profiles" json:"active_profiles"`
	DisabledProfiles int `db:"disabled_profiles" json:"disabled_profiles"`
	EmailsSentToday int `db:"emails_sent_today" json:"emails_sent_today"`
	FailedToday     int `db:"failed_today" json:"failed_today"`
}

type SMTPExportRow struct {
	Date        string  `db:"date" json:"date"`
	SMTPProfile string  `db:"smtp_profile" json:"smtp_profile"`
	EmailsSent  int     `db:"emails_sent" json:"emails_sent"`
	FailedEmails int    `db:"failed_emails" json:"failed_emails"`
	SuccessRate float64 `db:"success_rate" json:"success_rate"`
}

type SMTPExportRows []SMTPExportRow

// SMTPProfileWithStats extends SMTPProfile with live stat fields
type SMTPProfileWithStats struct {
	SMTPProfile
	SentToday   int        `db:"sent_today" json:"sent_today"`
	TotalSent   int        `db:"total_sent" json:"total_sent"`
	TotalFailed int        `db:"total_failed" json:"total_failed"`
	SuccessRate float64    `db:"success_rate" json:"success_rate"`
	LastSentAt  null.Time  `db:"last_sent_at" json:"last_sent_at"`
}

type SMTPProfilesWithStats []SMTPProfileWithStats
