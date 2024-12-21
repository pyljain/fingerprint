package routes

import (
	"fingerprint/internal/db"
	"net/http"

	"github.com/labstack/echo/v4"
)

// SetPreferencesForApp handles POST requests to update preferences for a specific application.
// It expects username and appName parameters in the URL path and a JSON body containing preferences.
// The handler removes username and appName from the preferences map if present.
// Returns an empty response on success or an error if the operation fails.
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

// SetPreferences handles POST requests to update all preferences for a user.
// Currently returns a placeholder response.
// TODO: Implement full functionality
func SetPreferences(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
