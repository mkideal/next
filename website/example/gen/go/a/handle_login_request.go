package a

import (
	"encoding/json"
	"net/http"
)

func handleLoginRequest(w http.ResponseWriter, r *http.Request) error {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	// TODO: handle request
	_ = req
	return nil
}
