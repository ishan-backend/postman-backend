package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ishan-backend/postman-backend/models/repo"
	"github.com/ishan-backend/postman-backend/models/reqResp"
)

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

/*

func (a *API) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id query param", http.StatusBadRequest)
		return
	}

	user, err := a.Services.Users.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	resp := reqResp.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}


func (a *API) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req reqResp.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.ID == "" {
		http.Error(w, "missing user id", http.StatusBadRequest)
		return
	}

	user := repo.User{
		ID:        req.ID,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	updated, err := a.Services.Users.UpdateUser(r.Context(), user)
	if err != nil {
		http.Error(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	resp := reqResp.UserResponse{
		ID:        updated.ID,
		Email:     updated.Email,
		FirstName: updated.FirstName,
		LastName:  updated.LastName,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

*/
