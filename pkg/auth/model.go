package auth

import (
	"context"
	"google.golang.org/grpc/credentials"
)

type TokenRequest struct {
	PodName string `json:"podName"`
	PodIp   string `json:"podIp"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type JWT struct {
	token string
}

func NewToken(token string) credentials.PerRPCCredentials {
	return JWT{token}
}

func (j JWT) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": j.token,
	}, nil
}

func (j JWT) RequireTransportSecurity() bool {
	return false
}
