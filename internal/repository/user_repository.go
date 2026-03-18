package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	rds "github.com/redis/go-redis/v9"

	"topikkhusus-methodtracker/internal/models"
)

const (
	userKeyPrefix = "user:"
	usersIndexKey = "users:index"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) error
	GetAllUsers(ctx context.Context) ([]models.User, error)
	GetUserByID(ctx context.Context, id string) (models.User, error)
	DeleteUser(ctx context.Context, id string) error
}

type RedisUserRepository struct {
	client  *rds.Client
	timeout time.Duration
}

func NewRedisUserRepository(client *rds.Client, timeout time.Duration) *RedisUserRepository {
	return &RedisUserRepository{client: client, timeout: timeout}
}

func (r *RedisUserRepository) CreateUser(ctx context.Context, user models.User) error {
	operationCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	userKey := userKey(user.ID)
	payload := map[string]interface{}{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"created_at": user.CreatedAt.Format(time.RFC3339Nano),
	}

	if err := r.client.HSet(operationCtx, userKey, payload).Err(); err != nil {
		return fmt.Errorf("hset user failed: %w", err)
	}

	if err := r.client.SAdd(operationCtx, usersIndexKey, user.ID).Err(); err != nil {
		return fmt.Errorf("sadd users index failed: %w", err)
	}

	return nil
}

func (r *RedisUserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	operationCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	ids, err := r.client.SMembers(operationCtx, usersIndexKey).Result()
	if err != nil {
		return nil, fmt.Errorf("smembers users index failed: %w", err)
	}

	users := make([]models.User, 0, len(ids))
	for _, id := range ids {
		user, userErr := r.GetUserByID(ctx, id)
		if userErr != nil {
			if errors.Is(userErr, ErrUserNotFound) {
				continue
			}

			return nil, fmt.Errorf("get user %s from index failed: %w", id, userErr)
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *RedisUserRepository) GetUserByID(ctx context.Context, id string) (models.User, error) {
	operationCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	result, err := r.client.HGetAll(operationCtx, userKey(id)).Result()
	if err != nil {
		return models.User{}, fmt.Errorf("hgetall user failed: %w", err)
	}

	if len(result) == 0 {
		return models.User{}, ErrUserNotFound
	}

	createdAt, err := time.Parse(time.RFC3339Nano, result["created_at"])
	if err != nil {
		return models.User{}, fmt.Errorf("parse created_at failed: %w", err)
	}

	return models.User{
		ID:        result["id"],
		Name:      result["name"],
		Email:     result["email"],
		CreatedAt: createdAt,
	}, nil
}

func (r *RedisUserRepository) DeleteUser(ctx context.Context, id string) error {
	operationCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	deleted, err := r.client.Del(operationCtx, userKey(id)).Result()
	if err != nil {
		return fmt.Errorf("delete user hash failed: %w", err)
	}

	if deleted == 0 {
		return ErrUserNotFound
	}

	if err := r.client.SRem(operationCtx, usersIndexKey, id).Err(); err != nil {
		return fmt.Errorf("srem users index failed: %w", err)
	}

	return nil
}

func userKey(id string) string {
	return userKeyPrefix + id
}
