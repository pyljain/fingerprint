package routes

import (
	"fingerprint/internal/db"
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetPreferencesForApp(c echo.Context) error {
	username := c.Param("username")
	appName := c.Param("appName")

	preferences := map[string]string{}
	err := c.Bind(&preferences)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid request body")
	}

	delete(preferences, "username")
	delete(preferences, "appName")

	// Upsert user preferences for app into MongoDB
	db := c.Get("db").(db.Database)
	err = db.UpsertPreferences(c.Request().Context(), username, appName, preferences)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, "")
}

func SetPreferences(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
