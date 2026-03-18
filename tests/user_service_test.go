package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"topikkhusus-methodtracker/internal/models"
	"topikkhusus-methodtracker/internal/repository"
	"topikkhusus-methodtracker/internal/services"
)

type mockUserRepository struct {
	users map[string]models.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{users: make(map[string]models.User)}
}

func (m *mockUserRepository) CreateUser(_ context.Context, user models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) GetAllUsers(_ context.Context) ([]models.User, error) {
	result := make([]models.User, 0, len(m.users))
	for _, user := range m.users {
		result = append(result, user)
	}

	return result, nil
}

func (m *mockUserRepository) GetUserByID(_ context.Context, id string) (models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return models.User{}, repository.ErrUserNotFound
	}

	return user, nil
}

func (m *mockUserRepository) DeleteUser(_ context.Context, id string) error {
	if _, ok := m.users[id]; !ok {
		return repository.ErrUserNotFound
	}

	delete(m.users, id)
	return nil
}

func TestCreateUserWithValidInput(t *testing.T) {
	repo := newMockUserRepository()
	service := services.NewUserService(repo)

	user, err := service.CreateUser(context.Background(), models.CreateUserRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID == "" {
		t.Fatalf("expected generated id")
	}

	if user.Name != "Alice" {
		t.Fatalf("expected name Alice, got %s", user.Name)
	}

	if user.Email != "alice@example.com" {
		t.Fatalf("expected email alice@example.com, got %s", user.Email)
	}

	if time.Since(user.CreatedAt) > time.Minute {
		t.Fatalf("expected recent created_at")
	}
}

func TestGetUserWithValidID(t *testing.T) {
	repo := newMockUserRepository()
	service := services.NewUserService(repo)

	created, err := service.CreateUser(context.Background(), models.CreateUserRequest{
		Name:  "Bob",
		Email: "bob@example.com",
	})
	if err != nil {
		t.Fatalf("expected no error while creating user, got %v", err)
	}

	fetched, err := service.GetUserByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected no error while fetching user, got %v", err)
	}

	if fetched.ID != created.ID {
		t.Fatalf("expected id %s, got %s", created.ID, fetched.ID)
	}
}

func TestGetUserWithNonExistentID(t *testing.T) {
	repo := newMockUserRepository()
	service := services.NewUserService(repo)

	_, err := service.GetUserByID(context.Background(), "missing-id")
	if !errors.Is(err, services.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
