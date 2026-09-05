package internal

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"trading-platform/pkg/authjwt"
	"trading-platform/pkg/web"

	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	users  *Repository
	tokens *authjwt.Tokens
}

func NewHandler(users *Repository, tokens *authjwt.Tokens) *Handler {
	return &Handler{users: users, tokens: tokens}
}

type registerRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	DisplayName *string `json:"displayName"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int64  `json:"expiresIn"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	email, details := normalizeEmail(req.Email)
	details = append(details, validatePassword(req.Password)...)
	displayName, nameDetails := normalizeDisplayName(req.DisplayName)
	details = append(details, nameDetails...)
	if len(details) > 0 {
		web.WriteValidation(w, details)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "Internal server error")
		return
	}

	created, err := h.users.Create(r.Context(), email, string(hash), displayName, "user")
	if errors.Is(err, ErrDuplicate) {
		web.WriteError(w, http.StatusConflict, web.CodeConflict, "Email already registered")
		return
	}
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "Internal server error")
		return
	}
	web.WriteJSON(w, http.StatusCreated, created.Public())
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	email, details := normalizeEmail(req.Email)
	if strings.TrimSpace(req.Password) == "" {
		details = append(details, web.FieldError{Field: "password", Message: "password is required"})
	}
	if len(details) > 0 {
		web.WriteValidation(w, details)
		return
	}

	u, err := h.users.GetByEmail(r.Context(), email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		web.WriteError(w, http.StatusUnauthorized, web.CodeUnauthorized, "Invalid email or password")
		return
	}

	token, err := h.tokens.Issue(u.ID, u.Email, u.Role)
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "Internal server error")
		return
	}
	web.WriteJSON(w, http.StatusOK, tokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.tokens.TTL().Seconds()),
	})
}

func normalizeEmail(raw string) (string, []web.FieldError) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", []web.FieldError{{Field: "email", Message: "email is required"}}
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "", []web.FieldError{{Field: "email", Message: "email is invalid"}}
	}
	return email, nil
}

func validatePassword(password string) []web.FieldError {
	if password == "" {
		return []web.FieldError{{Field: "password", Message: "password is required"}}
	}
	if len(password) < 8 || len(password) > 72 {
		return []web.FieldError{{Field: "password", Message: "password must be between 8 and 72 characters"}}
	}
	return nil
}

func normalizeDisplayName(raw *string) (*string, []web.FieldError) {
	if raw == nil {
		return nil, nil
	}
	name := strings.TrimSpace(*raw)
	if len(name) > 100 {
		return nil, []web.FieldError{{Field: "displayName", Message: "displayName must be at most 100 characters"}}
	}
	if name == "" {
		return nil, nil
	}
	return &name, nil
}
