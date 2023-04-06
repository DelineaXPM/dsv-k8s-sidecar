package secrets

import (
	"context"
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"
)

type SecretServer struct {
	client SecretClient
}

func NewSecretServer(client SecretClient) DsvServer {
	return &SecretServer{
		client,
	}
}

func (s *SecretServer) GetSecret(ctx context.Context, secret *Secret) (*Secret, error) {
	result, clientError := s.client.GetSecret(secret.Path)

	if clientError != nil {
		log.WithFields(log.Fields{
			"error":  clientError.Error.Error(),
			"status": clientError.Status,
		}).Error("Error connecting to API")
		return nil, errors.New("Error from API")
	}

	out, err := json.Marshal(result.Value)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"path":  result.Path,
		}).Error("unable to marshall value")
		return nil, errors.New("error unmarshalling data")
	}

	// Don't include attributes, can't be that flexible with grpc.
	resp := &Secret{
		Id:    result.ID,
		Path:  result.Path,
		Value: string(out),
	}

	return resp, nil
}
