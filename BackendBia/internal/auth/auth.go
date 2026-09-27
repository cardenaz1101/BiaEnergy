package auth

import (
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"github.com/biaenergy/backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

const defaultRole = "Energy Manager"

type User struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

type Credentials struct {
	Email    string
	Password string
	Name     string
}

type claims struct {
	Name string `json:"name"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type Service struct {
	secret      []byte
	credentials Credentials
	tokenTTL    time.Duration
}

func NewService(secret string, credentials Credentials, tokenTTL time.Duration) *Service {
	credentials.Email = normalizeEmail(credentials.Email)
	return &Service{secret: []byte(secret), credentials: credentials, tokenTTL: tokenTTL}
}

func (s *Service) Login(email, password string) (string, User, error) {
	emailMatches := subtle.ConstantTimeCompare([]byte(normalizeEmail(email)), []byte(s.credentials.Email)) == 1
	passwordMatches := subtle.ConstantTimeCompare([]byte(password), []byte(s.credentials.Password)) == 1
	if !emailMatches || !passwordMatches {
		return "", User{}, domain.ErrInvalidCredentials
	}
	user := User{Email: s.credentials.Email, Name: s.credentials.Name, Role: defaultRole}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Name: user.Name,
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
		},
	})
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", User{}, err
	}
	return signed, user, nil
}

func (s *Service) ParseToken(tokenString string) (User, error) {
	parsed := claims{}
	_, err := jwt.ParseWithClaims(tokenString, &parsed, func(*jwt.Token) (any, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return User{}, errors.Join(domain.ErrInvalidCredentials, err)
	}
	return User{Email: parsed.Subject, Name: parsed.Name, Role: parsed.Role}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
