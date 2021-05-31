// Package experiment is a test package.
package experiment

import (
	"time"

	"github.com/swaggest/swgen/internal/sample/experiment/variation"
)

type (
	// Entity defines experiment data.
	Entity struct {
		ID         int                `json:"id"`
		Variation  variation.Entity   `json:"variation"`
		Variations []variation.Entity `json:"variations"`
	}

	// Metadata defines experiment entity metadata.
	Metadata struct {
		SomePeriod Period `json:"allocation_period"`
	}

	// Period defines metadata period.
	Period struct {
		Start time.Time `json:"start" db:"-"`
		End   time.Time `json:"end" db:"-"`
	}

	// PostRequest is a test dummy.
	PostRequest struct {
		Country string `path:"country"`
		Data
	}

	// Data is a test dummy.
	Data struct {
		Metadata   Metadata        `json:"metadata"`
		Variations []VariationData `json:"variations"`
	}

	// VariationData is a test dummy.
	VariationData struct {
		SomeKey  string             `json:"some_key"`
		Metadata variation.Metadata `json:"metadata"`
	}
)
