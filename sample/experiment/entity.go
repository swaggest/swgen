package experiment

import (
	"github.com/swaggest/swgen/sample/experiment/variation"
	"time"
)

type (
	// Entity defines experiment data
	Entity struct {
		ID         int                `json:"id"`
		Variation  variation.Entity   `json:"variation"`
		Variations []variation.Entity `json:"variations"`
	}

	// Metadata defines experiment entity metadata
	Metadata struct {
		SomePeriod Period `json:"allocation_period"`
	}

	// Period defines metadata period
	Period struct {
		Start time.Time `json:"start" db:"-"`
		End   time.Time `json:"end" db:"-"`
	}

	PostRequest struct {
		Country string `path:"country"`
		Data    `json:"data"`
	}

	Data struct {
		Metadata   Metadata        `json:"metadata"`
		Variations []VariationData `json:"variations"`
	}

	VariationData struct {
		SomeKey  string             `json:"some_key"`
		Metadata variation.Metadata `json:"metadata"`
	}
)
