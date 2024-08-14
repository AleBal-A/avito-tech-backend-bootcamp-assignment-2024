package models

import "time"

type Flat struct {
	ID         int
	HouseID    int
	FlatNumber *int
	Price      int
	Rooms      int
	Status     string
	CreatedAt  time.Time
}
