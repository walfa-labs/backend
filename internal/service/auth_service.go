package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	adminRepo  port.AdminRepo
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(adminRepo port.AdminRepo, jwtSecret string, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		adminRepo:  adminRepo,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Login validates credentials and returns an access + refresh token pair.
func (s *AuthService) Login(ctx context.Context, username, password string) (*port.AuthTokens, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, domain.ErrUnauthorized
	}

	user, err := s.adminRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrUnauthorized
	}

	accessToken, err := s.issueToken(user.Username, "access", s.accessTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.issueToken(user.Username, "refresh", s.refreshTTL)
	if err != nil {
		return nil, err
	}

	return &port.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Refresh validates a refresh token and issues a new access token.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", domain.ErrUnauthorized
	}
	if claims.Type != "refresh" {
		return "", domain.ErrUnauthorized
	}
	return s.issueToken(claims.Subject, "access", s.accessTTL)
}

type jwtClaims struct {
	Username string `json:"username"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}

func (s *AuthService) issueToken(subject, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		Username: subject,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
