package db

import "context"

type Database interface {
	GetPreferences(ctx context.Context, username, appName string, fetchFromCache bool) (map[string]string, error)
	UpsertPreferences(ctx context.Context, username, appName string, preferences map[string]string) error
}
