package reqResp

import "go.mongodb.org/mongo-driver/bson/primitive"

// BulkUsersRequest represents the incoming payload for bulk user creation.
type BulkUsersRequest struct {
	Users []UserInput `json:"users"`
}

// UserInput is a single user payload in the request body.
type UserInput struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// BulkUsersResponse returns summary and inserted user IDs.
type BulkUsersResponse struct {
	InsertedCount int                  `json:"inserted_count"`
	InsertedIDs   []primitive.ObjectID `json:"inserted_ids"`
}
