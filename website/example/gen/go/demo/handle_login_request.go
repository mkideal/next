// Code generated by "next 0.0.4"; DO NOT EDIT.

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
	// TODO: handle request
	_ = req
	return nil
}