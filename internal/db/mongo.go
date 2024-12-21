package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewMongoAndRedisCache creates a new instance of MongoAndRedisCache with the provided connection strings.
// It establishes connections to both MongoDB and Redis servers and returns an error if either connection fails.
func NewMongoAndRedisCache(mongoConnectionString, redisConnectionString string) (*MongoAndRedisCache, error) {
	ctx := context.Background()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoConnectionString))
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisConnectionString,
	})

	status := redisClient.Ping(ctx)
	if status.Err() != nil {
		return nil, status.Err()
	}

	return &MongoAndRedisCache{
		mongoClient: mongoClient,
		redisClient: redisClient,
	}, nil
}

// UpsertPreferences updates or inserts user preferences for a specific application.
// It stores the preferences in both MongoDB and Redis for caching.
// The function returns an error if either the MongoDB operation or Redis caching fails.
func (m *MongoAndRedisCache) UpsertPreferences(ctx context.Context, username, appName string, preferences map[string]string) error {
	collection := m.mongoClient.Database("preferences").Collection("user_preferences")

	filter := bson.M{"username": username, "appName": appName}
	update := bson.M{"$set": bson.M{}}
	for prefKey, prefVal := range preferences {
		key := fmt.Sprintf("preferences.%s", prefKey)
		update["$set"].(bson.M)[key] = prefVal
	}

	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	preferences, err = m.GetPreferences(ctx, username, appName, false)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	redisResponse := m.redisClient.Set(ctx, fmt.Sprintf("%s-%s", username, appName), jsonData, 0)
	if redisResponse.Err() != nil {
		return redisResponse.Err()
	}

	return nil
}

// GetPreferences retrieves user preferences for a specific application.
// If fetchFromCache is true, it first attempts to retrieve from Redis cache.
// If cache miss or fetchFromCache is false, it retrieves from MongoDB and updates the cache.
// Returns the preferences map and any error encountered during the operation.
func (m *MongoAndRedisCache) GetPreferences(ctx context.Context, username, appName string, fetchFromCache bool) (map[string]string, error) {
	collection := m.mongoClient.Database("preferences").Collection("user_preferences")
	var preferences map[string]string

	if fetchFromCache {
		redisResult := m.redisClient.Get(ctx, fmt.Sprintf("%s-%s", username, appName))
		if redisResult.Err() == nil {
			res, err := redisResult.Result()
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal([]byte(res), &preferences)
			if err != nil {
				return nil, err
			}

			return preferences, nil
		}
	}

	filter := bson.M{"username": username, "appName": appName}
	result := collection.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var res PreferencesRecord
	err := result.Decode(&res)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal preferences: %w", err)
	}

	redisResponse := m.redisClient.Set(ctx, fmt.Sprintf("%s-%s", username, appName), jsonData, 0)
	if redisResponse.Err() != nil {
		return nil, redisResponse.Err()
	}

	return res.Preferences, nil
}

// PreferencesRecord represents the structure of a user preferences document in MongoDB.
type PreferencesRecord struct {
	AppName     string            `bson:"appName"`     // Name of the application
	Username    string            `bson:"username"`    // Username of the user
	Preferences map[string]string `bson:"preferences"` // Map of preference key-value pairs
}

// MongoAndRedisCache handles database operations with MongoDB and Redis caching.
type MongoAndRedisCache struct {
	mongoClient *mongo.Client // MongoDB client connection
	redisClient *redis.Client // Redis client connection
}
