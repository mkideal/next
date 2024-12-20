package demo

import (
	"encoding/json"
	"net/http"
)

func handleLoginRequest(w http.ResponseWriter, r *http.Request) error {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	_ = req
	return nil
}
