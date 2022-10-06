package auth

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

type authHandler struct {
	service AuthService
}

type AuthHandler interface {
	GetToken(w http.ResponseWriter, r *http.Request)
}

func NewAuthHandler(service AuthService) AuthHandler {
	return &authHandler{
		service,
	}
}

func (h *authHandler) GetToken(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest
	logrus.Info("Attempting login")
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Request is malformed", http.StatusBadRequest)
		return
	}

	resp := h.service.GetToken(&req)

	if resp == nil {
		http.Error(w, "Invalid Credentials", http.StatusForbidden)
		return
	}

	responseBody, err := json.Marshal(resp)
	if err != nil {
		logrus.Error("Error marshalling token")
		http.Error(w, "Unable to create token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseBody)
	if _, err = w.Write(responseBody); err != nil {
		logrus.Errorf("Error writing response: %s", err.Error())
	}
}
