package db

import "context"

// Database defines the interface for preference storage and retrieval operations.
// It abstracts the underlying database implementation, allowing for different storage solutions.
type Database interface {
	// GetPreferences retrieves user preferences for a specific application.
	// If fetchFromCache is true, it attempts to retrieve from cache first.
	GetPreferences(ctx context.Context, username, appName string, fetchFromCache bool) (map[string]string, error)

	// UpsertPreferences updates or inserts user preferences for a specific application.
	// It returns an error if the operation fails.
	UpsertPreferences(ctx context.Context, username, appName string, preferences map[string]string) error
}
