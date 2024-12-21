package routes

import (
	"fingerprint/internal/db"
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetPreferencesForApp handles GET requests to retrieve preferences for a specific application.
// It expects username and appName parameters in the URL path.
// Returns the preferences as JSON or an error if the retrieval fails.
func GetPreferencesForApp(c echo.Context) error {
	username := c.Param("username")
	appName := c.Param("appName")
	db := c.Get("db").(db.Database)
	preferences, err := db.GetPreferences(c.Request().Context(), username, appName, true)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, preferences)

}

// GetPreferences handles GET requests to retrieve all preferences for a user.
// Currently returns a placeholder response.
// TODO: Implement full functionality
func GetPreferences(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
