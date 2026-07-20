package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/knadh/koanf/providers/rawbytes"
	koanfjson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/listmonk/internal/messenger/email"
	"github.com/knadh/listmonk/internal/notifs"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetSMTPProfiles handles retrieval of all SMTP profiles.
func (a *App) GetSMTPProfiles(c echo.Context) error {
	profiles, err := a.core.QuerySMTPProfiles()
	if err != nil {
		return err
	}

	// Mask passwords.
	for i := range profiles {
		if profiles[i].Password != "" {
			profiles[i].Password = strings.Repeat(pwdMask, utf8.RuneCountInString(profiles[i].Password))
		}
	}

	return c.JSON(http.StatusOK, okResp{profiles})
}

// GetSMTPProfile handles retrieval of a single SMTP profile.
func (a *App) GetSMTPProfile(c echo.Context) error {
	id := getID(c)

	profile, err := a.core.GetSMTPProfile(id)
	if err != nil {
		return err
	}

	// Mask password.
	if profile.Password != "" {
		profile.Password = strings.Repeat(pwdMask, utf8.RuneCountInString(profile.Password))
	}

	return c.JSON(http.StatusOK, okResp{profile})
}

// CreateSMTPProfile handles creation of a new SMTP profile.
func (a *App) CreateSMTPProfile(c echo.Context) error {
	var o struct {
		Name       string `json:"name"`
		Host       string `json:"host"`
		Port       int    `json:"port"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		Encryption string `json:"encryption"`
		FromEmail  string `json:"from_email"`
		FromName   string `json:"from_name"`
		ReplyTo    string `json:"reply_to"`
		Enabled    bool   `json:"enabled"`
	}

	if err := c.Bind(&o); err != nil {
		return err
	}

	if o.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Profile name is required")
	}

	if o.Host == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "SMTP host is required")
	}

	if o.Port == 0 {
		o.Port = 587
	}

	if o.Encryption == "" {
		o.Encryption = "starttls"
	}

	profile, err := a.core.CreateSMTPProfile(models.SMTPProfile{
		UUID:       uuid.Must(uuid.NewV4()).String(),
		Name:       o.Name,
		Host:       o.Host,
		Port:       o.Port,
		Username:   o.Username,
		Password:   o.Password,
		Encryption: o.Encryption,
		FromEmail:  o.FromEmail,
		FromName:   o.FromName,
		ReplyTo:    o.ReplyTo,
		Enabled:    o.Enabled,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{profile})
}

// UpdateSMTPProfile handles updating an SMTP profile.
func (a *App) UpdateSMTPProfile(c echo.Context) error {
	id := getID(c)

	var o struct {
		Name       string `json:"name"`
		Host       string `json:"host"`
		Port       int    `json:"port"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		Encryption string `json:"encryption"`
		FromEmail  string `json:"from_email"`
		FromName   string `json:"from_name"`
		ReplyTo    string `json:"reply_to"`
		Enabled    bool   `json:"enabled"`
	}

	if err := c.Bind(&o); err != nil {
		return err
	}

	if o.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Profile name is required")
	}

	if o.Host == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "SMTP host is required")
	}

	if o.Port == 0 {
		o.Port = 587
	}

	// Get the existing profile to preserve password if not changed.
	existing, err := a.core.GetSMTPProfile(id)
	if err != nil {
		return err
	}

	password := o.Password
	if password == "" || password == strings.Repeat(pwdMask, utf8.RuneCountInString(existing.Password)) {
		password = existing.Password
	}

	profile, err := a.core.UpdateSMTPProfile(id, models.SMTPProfile{
		Name:       o.Name,
		Host:       o.Host,
		Port:       o.Port,
		Username:   o.Username,
		Password:   password,
		Encryption: o.Encryption,
		FromEmail:  o.FromEmail,
		FromName:   o.FromName,
		ReplyTo:    o.ReplyTo,
		Enabled:    o.Enabled,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{profile})
}

// DeleteSMTPProfile handles deletion of an SMTP profile.
func (a *App) DeleteSMTPProfile(c echo.Context) error {
	id := getID(c)

	if err := a.core.DeleteSMTPProfile(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{true})
}

// DuplicateSMTPProfile duplicates an existing SMTP profile.
func (a *App) DuplicateSMTPProfile(c echo.Context) error {
	id := getID(c)

	existing, err := a.core.GetSMTPProfile(id)
	if err != nil {
		return err
	}

	// Create a copy with a new name.
	newName := existing.Name + " (copy)"
	existing.Name = newName
	existing.ID = 0

	profile, err := a.core.CreateSMTPProfile(existing)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{profile})
}

// TestSMTPProfile tests an SMTP profile connection.
func (a *App) TestSMTPProfile(c echo.Context) error {
	// Read the raw body.
	reqBody, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "error reading request")
	}

	// Parse into koanf for proper handling.
	ko := koanf.New(".")
	if err := ko.Load(rawbytes.Provider(reqBody), koanfjson.Parser()); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "error parsing request")
	}

	var profile models.SMTPProfile
	if err := ko.UnmarshalWithConf("", &profile, koanf.UnmarshalConf{Tag: "json"}); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "error parsing request")
	}

	// Map encryption to TLS type.
	tlsType := "STARTTLS"
	switch profile.Encryption {
	case "ssl_tls":
		tlsType = "TLS"
	case "none":
		tlsType = "none"
	default:
		tlsType = "STARTTLS"
	}

	// Build SMTP server config.
	srv := email.Server{
		Username:     profile.Username,
		Password:     profile.Password,
		AuthProtocol: "login",
		TLSType:      tlsType,
	}

	srv.Opt.Host = profile.Host
	srv.Opt.Port = profile.Port
	srv.MaxConns = 1
	srv.Opt.PoolWaitTimeout = time.Second * 5
	srv.IdleTimeout = (time.Second * 2).String()
	srv.WaitTimeout = (time.Second * 2).String()

	msgr, err := email.New("test", srv)
	if err != nil {
		return c.JSON(http.StatusOK, okResp{map[string]string{
			"status":  "error",
			"message": err.Error(),
		}})
	}

	// Render test email template.
	var b bytes.Buffer
	if err := notifs.Tpls.ExecuteTemplate(&b, "smtp-test", nil); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error rendering test template")
	}

	m := models.Message{}
	m.From = profile.FromEmail
	if m.From == "" {
		m.From = "test@localhost"
	}
	m.Subject = "SMTP Profile Test"
	m.Body = b.Bytes()

	to := ko.String("email")
	if to != "" {
		m.To = []string{to}
	}

	if err := msgr.Push(m); err != nil {
		return c.JSON(http.StatusOK, okResp{map[string]string{
			"status":  "error",
			"message": err.Error(),
		}})
	}

	return c.JSON(http.StatusOK, okResp{map[string]string{
		"status":  "success",
		"message": "Connection successful",
	}})
}
