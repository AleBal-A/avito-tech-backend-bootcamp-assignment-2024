package models

type Flat struct {
	ID          int
	HouseID     int
	FlatNumber  *int
	Price       int
	Rooms       int
	Status      string
	ModeratorID *string
}
