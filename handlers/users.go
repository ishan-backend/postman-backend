package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ishan-backend/postman-backend/models/repo"
	"github.com/ishan-backend/postman-backend/models/reqResp"
)

// BulkCreateUsers handles POST /users/bulk to insert many users atomically.
func (a *API) BulkCreateUsers(w http.ResponseWriter, r *http.Request) {
	var req reqResp.BulkUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid json"))
		return
	}
	if len(req.Users) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("no users provided"))
		return
	}

	users := make([]repo.User, 0, len(req.Users))
	for _, in := range req.Users {
		users = append(users, repo.User{
			Email:     in.Email,
			FirstName: in.FirstName,
			LastName:  in.LastName,
		})
	}

	ids, err := a.Services.Users.BulkCreateUsers(r.Context(), users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to create users"))
		return
	}

	resp := reqResp.BulkUsersResponse{InsertedCount: len(ids), InsertedIDs: ids}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
