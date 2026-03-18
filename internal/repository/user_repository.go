package repository

import (
	"context"

	"topikkhusus-methodtracker/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) error
	GetAllUsers(ctx context.Context) ([]models.User, error)
	GetUserByID(ctx context.Context, id string) (models.User, error)
	DeleteUser(ctx context.Context, id string) error
}
