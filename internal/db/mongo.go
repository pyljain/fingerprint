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

type PreferencesRecord struct {
	AppName     string            `bson:"appName"`
	Username    string            `bson:"username"`
	Preferences map[string]string `bson:"preferences"`
}
type MongoAndRedisCache struct {
	mongoClient *mongo.Client
	redisClient *redis.Client
}
