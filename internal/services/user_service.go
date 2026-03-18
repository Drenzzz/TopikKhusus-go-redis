package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"topikkhusus-methodtracker/internal/models"
	"topikkhusus-methodtracker/internal/repository"
)

var (
	ErrNotFound     = errors.New("user not found")
	ErrInvalidInput = errors.New("invalid user input")
)

type UserService interface {
	CreateUser(ctx context.Context, request models.CreateUserRequest) (models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
	GetUserByID(ctx context.Context, id string) (models.User, error)
	DeleteUser(ctx context.Context, id string) error
}

type userService struct {
	repository repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repository: repo}
}

func (s *userService) CreateUser(ctx context.Context, request models.CreateUserRequest) (models.User, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Email = strings.TrimSpace(request.Email)

	if request.Name == "" || request.Email == "" {
		return models.User{}, ErrInvalidInput
	}

	return models.User{}, fmt.Errorf("create user is not implemented")
}

func (s *userService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.repository.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all users failed: %w", err)
	}

	return users, nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (models.User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return models.User{}, ErrInvalidInput
	}

	user, err := s.repository.GetUserByID(ctx, id)
	if err != nil {
		return models.User{}, fmt.Errorf("get user by id failed: %w", err)
	}

	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidInput
	}

	if err := s.repository.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}

	return nil
}
