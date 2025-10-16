package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ishan-backend/postman-backend/mocks"
	"github.com/ishan-backend/postman-backend/models/repo"
	"github.com/ishan-backend/postman-backend/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserService_BulkCreateUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockUsersRepoInterface(ctrl)

	// Initialize service with mock repository
	userService := &service.UserService{UsersRepo: mockRepo}

	// Predefined object IDs
	id1 := primitive.NewObjectID()
	id2 := primitive.NewObjectID()

	// Test cases
	tests := []struct {
		name          string
		inputUsers    []repo.User
		mockSetup     func()
		expectedIDs   []primitive.ObjectID
		expectedError error
	}{
		{
			name:          "empty input - returns nil",
			inputUsers:    []repo.User{},
			mockSetup:     func() {}, // no repo call expected
			expectedIDs:   nil,
			expectedError: nil,
		},
		{
			name: "successful bulk insert",
			inputUsers: []repo.User{
				{FirstName: "Alice", Email: "alice@test.com"},
				{FirstName: "Bob", Email: "bob@test.com"},
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					BulkInsert(ctx, []repo.User{
						{FirstName: "Alice", Email: "alice@test.com"},
						{FirstName: "Bob", Email: "bob@test.com"},
					}).
					Return([]primitive.ObjectID{id1, id2}, nil).
					Times(1)
			},
			expectedIDs:   []primitive.ObjectID{id1, id2},
			expectedError: nil,
		},
		{
			name: "repository error",
			inputUsers: []repo.User{
				{FirstName: "Charlie", Email: "charlie@test.com"},
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					BulkInsert(ctx, []repo.User{
						{FirstName: "Charlie", Email: "charlie@test.com"},
					}).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			expectedIDs:   nil,
			expectedError: errors.New("db error"),
		},
	}

	// Loop through test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock behavior
			tc.mockSetup()

			// Call service method
			actualIDs, actualErr := userService.BulkCreateUsers(ctx, tc.inputUsers)

			// Assert results
			if tc.expectedError != nil {
				assert.EqualError(t, actualErr, tc.expectedError.Error())
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedIDs, actualIDs)
		})
	}
}
