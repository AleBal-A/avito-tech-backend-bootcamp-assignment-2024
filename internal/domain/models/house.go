package models

import "time"

type House struct {
	ID            int
	Address       string
	YearBuilt     int
	Builder       *string
	CreatedAt     time.Time
	LastFlatAdded *time.Time
}
