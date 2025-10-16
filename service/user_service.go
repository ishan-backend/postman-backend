package service

import (
	"context"

	"github.com/ishan-backend/postman-backend/models/repo"
	"github.com/ishan-backend/postman-backend/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserServiceInterface defines business operations for users.
type UserServiceInterface interface {
	BulkCreateUsers(ctx context.Context, users []repo.User) ([]primitive.ObjectID, error)
}

// UserService implements user business logic.
type UserService struct {
	usersRepo repository.UsersRepoInterface
}

// NewUserService constructs a UserService.
func NewUserService(repos *repository.Repositories) *UserService {
	return &UserService{usersRepo: repository.NewUsersRepository(repos)}
}

// BulkCreateUsers validates and forwards to repository.
func (s *UserService) BulkCreateUsers(ctx context.Context, users []repo.User) ([]primitive.ObjectID, error) {
	if len(users) == 0 {
		return nil, nil
	}
	return s.usersRepo.BulkInsert(ctx, users)
}
