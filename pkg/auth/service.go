package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/pods"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authService struct {
	registry pods.PodRegistry
	secret   string
}

type AuthService interface {
	GetToken(request *TokenRequest) *TokenResponse
	GetUnaryInterceptor() grpc.ServerOption
}

func NewAuthService(secret string, registry pods.PodRegistry) AuthService {
	return &authService{
		registry,
		secret,
	}
}

func (s *authService) GetToken(request *TokenRequest) *TokenResponse {
	key := []byte(s.secret)
	logrus.Info("podname " + request.PodName)
	pod := s.registry.Get(request.PodName)
	if pod == nil {
		logrus.Errorf("pod is nil")
	}
	if pod == nil || *pod.Status.PodIP != request.PodIp {
		return nil
	}

	/* Create the token */
	token := jwt.New(jwt.SigningMethodHS256)

	/* Create a map to store our claims */
	claims := token.Claims.(jwt.MapClaims)

	/* Set token claims */
	claims["sub"] = *pod.Metadata.Uid
	claims["name"] = *pod.Metadata.Name
	claims["type"] = "pod"
	claims["exp"] = time.Now().Add(time.Hour * 2400).Unix() // Long Term token.

	/* Sign the token with our secret */
	tokenString, _ := token.SignedString(key)

	return &TokenResponse{
		Token: tokenString,
	}
}

func (s *authService) GetUnaryInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(s.unaryInterceptor)
}

func extractHeader(ctx context.Context, header string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no headers in request")
	}

	authHeaders, ok := md[header]
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no header in request")
	}

	if len(authHeaders) != 1 {
		return "", status.Error(codes.Unauthenticated, "more than 1 header in request")
	}

	return authHeaders[0], nil
}

func purgeHeader(ctx context.Context, header string) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	mdCopy := md.Copy()
	mdCopy[header] = nil
	return metadata.NewIncomingContext(ctx, mdCopy)
}

func (s *authService) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, err := extractHeader(ctx, "authorization")
	if err != nil {
		return nil, err
	}

	_, err = jwt.Parse(md, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.secret), nil
	})

	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}

	ctx = purgeHeader(ctx, "authorization")
	return handler(ctx, req)
}
