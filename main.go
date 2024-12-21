package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"fingerprint/internal/db"
	"fingerprint/internal/routes"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	mongoConnectionString := os.Getenv("MONGO_CONNECTION_STRING")
	redisConnectionString := os.Getenv("REDIS_CONNECTION_STRING")

	db, err := db.NewMongoAndRedisCache(mongoConnectionString, redisConnectionString)
	if err != nil {
		slog.Error("failed to create database", "error", err)
		os.Exit(1)
	}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", db)
			return next(c)
		}
	})

	e.GET("/api/v1/preferences/users/:username/applications/:appName", routes.GetPreferencesForApp)
	e.GET("/api/v1/preferences/users/:username", routes.GetPreferences)
	e.POST("/api/v1/preferences/users/:username/applications/:appName", routes.SetPreferencesForApp)
	e.POST("/api/v1/preferences/users/:username", routes.SetPreferences)

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
