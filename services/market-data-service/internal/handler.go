package internal

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"trading-platform/pkg/web"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var symbolPattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9.-]{0,31}$`)

var allowedAssetClass = map[string]struct{}{
	"stock":  {},
	"crypto": {},
	"forex":  {},
	"other":  {},
}

type Handler struct {
	instruments *Repository
}

func NewHandler(instruments *Repository) *Handler {
	return &Handler{instruments: instruments}
}

type instrumentRequest struct {
	Symbol     string `json:"symbol"`
	Name       string `json:"name"`
	AssetClass string `json:"assetClass"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize, details := web.ParsePage(r)
	if len(details) > 0 {
		web.WriteValidation(w, details)
		return
	}
	items, total, err := h.instruments.List(r.Context(), page, pageSize)
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "Internal server error")
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Page[Instrument]{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	item, err := h.instruments.GetByID(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		web.WriteError(w, http.StatusNotFound, web.CodeNotFound, "Instrument not found")
		return
	}
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "Internal server error")
		return
	}
	web.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeInstrument(w, r)
	if !ok {
		return
	}
	item, err := h.instruments.Create(r.Context(), req.Symbol, req.Name, req.AssetClass)
	if errors.Is(err, ErrDuplicate) {
		web.WriteError(w, http.StatusConflict, web.CodeConflict, "Instrument symbol already exists")
		return
	}
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "Internal server error")
		return
	}
	web.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	req, ok := decodeInstrument(w, r)
	if !ok {
		return
	}
	item, err := h.instruments.Update(r.Context(), id, req.Symbol, req.Name, req.AssetClass)
	if errors.Is(err, ErrNotFound) {
		web.WriteError(w, http.StatusNotFound, web.CodeNotFound, "Instrument not found")
		return
	}
	if errors.Is(err, ErrDuplicate) {
		web.WriteError(w, http.StatusConflict, web.CodeConflict, "Instrument symbol already exists")
		return
	}
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "Internal server error")
		return
	}
	web.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.instruments.Delete(r.Context(), id); errors.Is(err, ErrNotFound) {
		web.WriteError(w, http.StatusNotFound, web.CodeNotFound, "Instrument not found")
		return
	} else if err != nil {
		web.WriteError(w, http.StatusInternalServerError, web.CodeInternal, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteValidation(w, []web.FieldError{{Field: "id", Message: "id must be a UUID"}})
		return uuid.UUID{}, false
	}
	return id, true
}

func decodeInstrument(w http.ResponseWriter, r *http.Request) (instrumentRequest, bool) {
	var req instrumentRequest
	if !web.DecodeJSON(w, r, &req) {
		return req, false
	}
	var details []web.FieldError

	req.Symbol = strings.ToUpper(strings.TrimSpace(req.Symbol))
	if req.Symbol == "" {
		details = append(details, web.FieldError{Field: "symbol", Message: "symbol is required"})
	} else if !symbolPattern.MatchString(req.Symbol) {
		details = append(details, web.FieldError{Field: "symbol", Message: "symbol must be 1-32 chars: letters, digits, dot or hyphen"})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		details = append(details, web.FieldError{Field: "name", Message: "name is required"})
	} else if len(req.Name) > 200 {
		details = append(details, web.FieldError{Field: "name", Message: "name must be at most 200 characters"})
	}

	req.AssetClass = strings.ToLower(strings.TrimSpace(req.AssetClass))
	if req.AssetClass == "" {
		details = append(details, web.FieldError{Field: "assetClass", Message: "assetClass is required"})
	} else if _, ok := allowedAssetClass[req.AssetClass]; !ok {
		details = append(details, web.FieldError{Field: "assetClass", Message: "assetClass must be stock, crypto, forex or other"})
	}

	if len(details) > 0 {
		web.WriteValidation(w, details)
		return req, false
	}
	return req, true
}
