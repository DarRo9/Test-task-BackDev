package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/DarRo9/Test-task-BackDev/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	usersCollection = "users"
	name            = "name"
	rToken          = "refresh_token"
	createdTime     = "created_time"
)

type RefreshRepo struct {
	db *mongo.Collection
}

type Storage struct {
	db *mongo.Database
}

func (r *RefreshRepo) DeleteToken(ctx context.Context, refreshToken string) error {
	const op = "storage.mongodb.DeleteToken"

	filter := bson.M{rToken: refreshToken}

	if _, err := r.db.DeleteOne(ctx, filter); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RefreshRepo) InsertToken(ctx context.Context, userName string, refreshToken string, timeNow time.Time) error {
	const op = "storage.mongodb.InsertToken"

	if _, err := r.db.InsertOne(ctx, models.User{
		Name:         userName,
		RefreshToken: refreshToken,
		CreatedTime:  timeNow,
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RefreshRepo) DeleteTokensByUser(ctx context.Context, userName string) error {
	const op = "storage.mongodb.DeleteTokensByUser"

	filter := bson.M{name: userName}

	if _, err := r.db.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RefreshRepo) GetTokenByUser(ctx context.Context, userName string) (string, error) {
	const op = "storage.mongodb.GetTokenByUser"

	filter := bson.M{name: userName}

	var user models.User
	err := r.db.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return user.RefreshToken, nil
}

func (r *RefreshRepo) SelectToken(ctx context.Context, oldRefreshToken string, CreateObjectRefreshToken string, userName string, timeNow time.Time) error {
	const op = "storage.mongodb.SelectToken"

	if err := r.DeleteToken(ctx, oldRefreshToken); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := r.InsertToken(ctx, userName, CreateObjectRefreshToken, timeNow); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RefreshRepo) CountTokens(ctx context.Context, userName string) (int64, error) {
	const op = "storage.mongodb.CountTokens"

	filter := bson.M{name: userName}

	count, err := r.db.CountDocuments(context.Background(), filter)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return count, nil
}

func (r *RefreshRepo) GetTime(ctx context.Context, refreshToken string, userName string) (time.Time, error) {
	const op = "storage.mongodb.GetTime"

	filter := bson.M{rToken: refreshToken, name: userName}

	var user models.User
	err := r.db.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	return user.CreatedTime, nil
}

func CreateObjectStorage(client *mongo.Client, database string) *Storage {
	return &Storage{db: client.Database(database)}
}

func (s *Storage) CreateObjectRefreshRepo() *RefreshRepo {
	return &RefreshRepo{
		db: s.db.Collection(usersCollection),
	}
}
