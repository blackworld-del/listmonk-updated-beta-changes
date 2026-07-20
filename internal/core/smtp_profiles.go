package core

import (
	"database/sql"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

const smtpProfileNameDefault = "Default SMTP"

func (c *Core) QuerySMTPProfiles() (models.SMTPProfiles, error) {
	var out models.SMTPProfiles
	if err := c.q.QuerySMTPProfiles.Select(&out); err != nil {
		c.log.Printf("error fetching SMTP profiles: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "SMTP profiles", "error", pqErrMsg(err)))
	}

	return out, nil
}

func (c *Core) GetSMTPProfile(id int) (models.SMTPProfile, error) {
	var out models.SMTPProfile
	if err := c.q.GetSMTPProfile.Get(&out, id, nil); err != nil {
		if err == sql.ErrNoRows {
			return out, echo.NewHTTPError(http.StatusBadRequest,
				c.i18n.Ts("globals.messages.notFound", "name", "SMTP profile"))
		}

		c.log.Printf("error fetching SMTP profile: %v", err)
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "SMTP profile", "error", pqErrMsg(err)))
	}

	return out, nil
}

func (c *Core) CreateSMTPProfile(o models.SMTPProfile) (models.SMTPProfile, error) {
	// Validate name uniqueness.
	var exists int
	if err := c.q.GetSMTPProfileByName.Get(&exists, o.Name); err == nil && exists > 0 {
		return models.SMTPProfile{}, echo.NewHTTPError(http.StatusBadRequest,
			"SMTP profile with this name already exists")
	}

	uu, err := uuid.NewV4()
	if err != nil {
		c.log.Printf("error generating UUID: %v", err)
		return models.SMTPProfile{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUUID", "error", err.Error()))
	}

	var newID int
	if err := c.q.CreateSMTPProfile.Get(&newID,
		uu.String(),
		o.Name,
		o.Host,
		o.Port,
		o.Username,
		o.Password,
		o.Encryption,
		o.FromEmail,
		o.FromName,
		o.ReplyTo,
		o.Enabled,
	); err != nil {
		c.log.Printf("error creating SMTP profile: %v", err)
		return models.SMTPProfile{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "SMTP profile", "error", pqErrMsg(err)))
	}

	return c.GetSMTPProfile(newID)
}

func (c *Core) UpdateSMTPProfile(id int, o models.SMTPProfile) (models.SMTPProfile, error) {
	_, err := c.q.UpdateSMTPProfile.Exec(id,
		o.Name,
		o.Host,
		o.Port,
		o.Username,
		o.Password,
		o.Encryption,
		o.FromEmail,
		o.FromName,
		o.ReplyTo,
		o.Enabled,
	)
	if err != nil {
		c.log.Printf("error updating SMTP profile: %v", err)
		return models.SMTPProfile{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "SMTP profile", "error", pqErrMsg(err)))
	}

	return c.GetSMTPProfile(id)
}

func (c *Core) DeleteSMTPProfile(id int) error {
	// Check if the profile is used by any campaigns.
	var campaignCount int
	if err := c.db.Get(&campaignCount, `SELECT COUNT(*) FROM campaigns WHERE smtp_profile_id = $1`, id); err != nil {
		c.log.Printf("error checking SMTP profile usage: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "campaigns", "error", pqErrMsg(err)))
	}

	if campaignCount > 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			"Cannot delete SMTP profile that is assigned to one or more campaigns")
	}

	_, err := c.q.DeleteSMTPProfile.Exec(id)
	if err != nil {
		c.log.Printf("error deleting SMTP profile: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "SMTP profile", "error", pqErrMsg(err)))
	}

	return nil
}
