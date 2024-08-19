package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Users struct {
	Name string `json:"name"`
}
type CreatedUser struct {
	ID         uuid.UUID `json:"id"`
	Created_At time.Time `json:"created_at"`
	Updated_At time.Time `json:"updated_at"`
	Name       string    `json:"name"`
}

func decodeUsers(r *http.Request) (Users, error) {
	newUser := Users{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&newUser)
	if err != nil {
		return Users{}, err
	}

	return newUser, nil
}
