package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/labstack/echo/v4"
)

type smtpStatsResp struct {
	Stats     interface{} `json:"stats"`
	Daily     interface{} `json:"daily"`
	Campaigns interface{} `json:"campaigns"`
	Activity  interface{} `json:"activity"`
}

func (a *App) GetSMTPProfileStats(c echo.Context) error {
	id := getID(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	stats, daily, campaigns, activity, err := a.core.GetSMTPProfileStats(id, page, perPage)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{smtpStatsResp{
		Stats:     stats,
		Daily:     daily,
		Campaigns: campaigns,
		Activity:  activity,
	}})
}

func (a *App) GetSMTPOverview(c echo.Context) error {
	overview, err := a.core.GetSMTPOverview()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{overview})
}

func (a *App) GetSMTPDashboardStats(c echo.Context) error {
	stats, err := a.core.GetSMTPDashboardStats()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{stats})
}

func (a *App) ExportSMTPStats(c echo.Context) error {
	data, err := a.core.ExportSMTPStats()
	if err != nil {
		return err
	}

	format := c.QueryParam("format")
	if format == "csv" {
		c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
		c.Response().Header().Set("Content-Disposition", "attachment; filename=smtp-stats.csv")
		return c.Blob(http.StatusOK, "text/csv; charset=utf-8", data)
	}

	return c.JSON(http.StatusOK, okResp{string(data)})
}

func (a *App) QuerySMTPProfilesWithStats(c echo.Context) error {
	orderBy := c.QueryParam("order_by")
	order := c.QueryParam("order")

	profiles, err := a.core.QuerySMTPProfilesWithStats(orderBy, order)
	if err != nil {
		return err
	}

	// Mask passwords
	for i := range profiles {
		if profiles[i].Password != "" {
			profiles[i].Password = strings.Repeat(pwdMask, utf8.RuneCountInString(profiles[i].Password))
		}
	}

	return c.JSON(http.StatusOK, okResp{profiles})
}

func (a *App) RecordSMTPActivity(c echo.Context) error {
	id := getID(c)

	var o struct {
		EventType string `json:"event_type"`
		Message   string `json:"message"`
	}
	if err := c.Bind(&o); err != nil {
		return err
	}

	if err := a.core.InsertSMTPActivity(id, o.EventType, o.Message); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{true})
}
