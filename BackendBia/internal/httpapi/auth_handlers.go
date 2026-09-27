package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/biaenergy/backend/internal/auth"
)

type contextKey string

const userContextKey contextKey = "user"

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string    `json:"token"`
	User  auth.User `json:"user"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	token, user, err := s.Auth.Login(request.Email, request.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{Token: token, User: user})
}

func (s *Server) currentUser(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, userFromContext(r.Context()))
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !found {
			writeError(w, http.StatusUnauthorized, "no autenticado")
			return
		}
		user, err := s.Auth.ParseToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "no autenticado")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
	})
}

func userFromContext(ctx context.Context) auth.User {
	user, _ := ctx.Value(userContextKey).(auth.User)
	return user
}
