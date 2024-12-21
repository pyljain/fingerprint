package routes

import (
	"fingerprint/internal/db"
	"net/http"

	"github.com/labstack/echo/v4"
)

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

func GetPreferences(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
