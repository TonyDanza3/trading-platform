package internal

import (
	"time"

	"github.com/google/uuid"
)

type Instrument struct {
	ID         uuid.UUID `json:"id"`
	Symbol     string    `json:"symbol"`
	Name       string    `json:"name"`
	AssetClass string    `json:"assetClass"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
