package dbschemas

// SavedLocation : saved location
type SavedLocation struct {
	WorldId      [16]byte
	SavedBy      string
	LocationName string
	XPos         float64
	YPos         float64
	ZPos         float64
}
