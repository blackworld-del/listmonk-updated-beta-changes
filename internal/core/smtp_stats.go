package core

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

func (c *Core) GetSMTPProfileStats(id int, page, perPage int) (*models.SMTPProfileStats, []models.SMTPDailyStat, []models.SMTPRecentCampaign, []models.SMTPActivity, error) {
	var stats models.SMTPProfileStats
	if err := c.q.GetSMTPProfileStats.Get(&stats, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil, nil, echo.NewHTTPError(http.StatusBadRequest,
				c.i18n.Ts("globals.messages.notFound", "name", "SMTP profile"))
		}
		c.log.Printf("error fetching SMTP profile stats: %v", err)
		return nil, nil, nil, nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "SMTP stats", "error", pqErrMsg(err)))
	}

	var daily []models.SMTPDailyStat
	if err := c.q.GetSMTPProfileDailyStats.Select(&daily, id); err != nil {
		c.log.Printf("error fetching daily stats: %v", err)
	}

	offset := (page - 1) * perPage
	var campaigns []models.SMTPRecentCampaign
	if err := c.q.GetSMTPProfileRecentCampaigns.Select(&campaigns, id, perPage, offset); err != nil {
		c.log.Printf("error fetching recent campaigns: %v", err)
	}

	var activity []models.SMTPActivity
	if err := c.q.GetSMTPProfileActivity.Select(&activity, id, perPage, offset); err != nil {
		c.log.Printf("error fetching activity: %v", err)
	}

	return &stats, daily, campaigns, activity, nil
}

func (c *Core) GetSMTPOverview() ([]models.SMTPOverview, error) {
	var out []models.SMTPOverview
	if err := c.q.GetSMTPOverview.Select(&out); err != nil {
		c.log.Printf("error fetching SMTP overview: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "SMTP overview", "error", pqErrMsg(err)))
	}
	return out, nil
}

func (c *Core) GetSMTPDashboardStats() (*models.SMTPDashboardStats, error) {
	var out models.SMTPDashboardStats
	if err := c.q.GetSMTPDashboardStats.Get(&out); err != nil {
		c.log.Printf("error fetching SMTP dashboard stats: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "SMTP dashboard stats", "error", pqErrMsg(err)))
	}
	return &out, nil
}

func (c *Core) ExportSMTPStats() ([]byte, error) {
	var rows models.SMTPExportRows
	if err := c.q.ExportSMTPStats.Select(&rows); err != nil {
		c.log.Printf("error exporting SMTP stats: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "SMTP export", "error", pqErrMsg(err)))
	}

	var b strings.Builder
	writer := csv.NewWriter(&b)
	writer.Write([]string{"Date", "SMTP Profile", "Emails Sent", "Failed Emails", "Success Rate"})
	for _, r := range rows {
		writer.Write([]string{
			r.Date,
			r.SMTPProfile,
			fmt.Sprintf("%d", r.EmailsSent),
			fmt.Sprintf("%d", r.FailedEmails),
			fmt.Sprintf("%.1f%%", r.SuccessRate),
		})
	}
	writer.Flush()

	return []byte(b.String()), nil
}

func (c *Core) QuerySMTPProfilesWithStats(orderBy, order string) (models.SMTPProfilesWithStats, error) {
	sortFields := []string{"name", "host", "username", "sent_today", "total_sent", "total_failed", "success_rate", "last_sent_at"}
	if !strSliceContains(orderBy, sortFields) {
		orderBy = "name"
	}
	if order != SortAsc && order != SortDesc {
		order = SortAsc
	}

	query := strings.ReplaceAll(c.q.QuerySMTPProfilesStats, "%order%", orderBy+" "+order)

	var out models.SMTPProfilesWithStats
	if err := c.db.Select(&out, query); err != nil {
		c.log.Printf("error querying SMTP profiles with stats: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "SMTP profiles", "error", pqErrMsg(err)))
	}

	return out, nil
}

func (c *Core) InsertSMTPActivity(smtpProfileID int, eventType, message string) error {
	_, err := c.q.InsertSMTPActivity.Exec(smtpProfileID, eventType, message)
	if err != nil {
		c.log.Printf("error inserting SMTP activity: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "SMTP activity", "error", pqErrMsg(err)))
	}
	return nil
}
