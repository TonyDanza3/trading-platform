package internal

import (
	"errors"
	"net/http"
	"strings"

	"trading-platform/pkg/web"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	users *Repository
}

func NewHandler(users *Repository) *Handler {
	return &Handler{users: users}
}

type patchMeRequest struct {
	DisplayName *string `json:"displayName"`
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	p, ok := web.PrincipalFrom(r.Context())
	if !ok {
		web.WriteError(w, http.StatusUnauthorized, web.CodeUnauthorized, "Missing or invalid access token")
		return
	}
	u, err := h.users.GetByID(r.Context(), p.ID)
	if errors.Is(err, ErrNotFound) {
		web.WriteError(w, http.StatusUnauthorized, web.CodeUnauthorized, "Missing or invalid access token")
		return
	}
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "Internal server error")
		return
	}
	web.WriteJSON(w, http.StatusOK, u.Public())
}

func (h *Handler) PatchMe(w http.ResponseWriter, r *http.Request) {
	p, ok := web.PrincipalFrom(r.Context())
	if !ok {
		web.WriteError(w, http.StatusUnauthorized, web.CodeUnauthorized, "Missing or invalid access token")
		return
	}
	var req patchMeRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	if req.DisplayName == nil {
		web.WriteValidation(w, []web.FieldError{{Field: "displayName", Message: "displayName is required"}})
		return
	}
	name := strings.TrimSpace(*req.DisplayName)
	if len(name) > 100 {
		web.WriteValidation(w, []web.FieldError{{Field: "displayName", Message: "displayName must be at most 100 characters"}})
		return
	}
	var stored *string
	if name != "" {
		stored = &name
	}
	u, err := h.users.UpdateDisplayName(r.Context(), p.ID, stored)
	if errors.Is(err, ErrNotFound) {
		web.WriteError(w, http.StatusUnauthorized, web.CodeUnauthorized, "Missing or invalid access token")
		return
	}
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "Internal server error")
		return
	}
	web.WriteJSON(w, http.StatusOK, u.Public())
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize, details := web.ParsePage(r)
	if len(details) > 0 {
		web.WriteValidation(w, details)
		return
	}
	items, total, err := h.users.List(r.Context(), page, pageSize)
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "Internal server error")
		return
	}
	public := make([]PublicUser, 0, len(items))
	for _, item := range items {
		public = append(public, item.Public())
	}
	web.WriteJSON(w, http.StatusOK, web.Page[PublicUser]{
		Items:    public,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	p, ok := web.PrincipalFrom(r.Context())
	if !ok {
		web.WriteError(w, http.StatusUnauthorized, web.CodeUnauthorized, "Missing or invalid access token")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteValidation(w, []web.FieldError{{Field: "id", Message: "id must be a UUID"}})
		return
	}
	if id == p.ID {
		web.WriteError(w, http.StatusConflict, web.CodeConflict, "Cannot delete yourself")
		return
	}
	if err := h.users.Delete(r.Context(), id); errors.Is(err, ErrNotFound) {
		web.WriteError(w, http.StatusNotFound, web.CodeNotFound, "User not found")
		return
	} else if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
