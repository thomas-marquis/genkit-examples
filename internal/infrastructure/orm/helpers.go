package orm

import (
	"fmt"
	"strconv"
)

// Helper to convert uint ID to string
func idToString(id uint) string {
	return fmt.Sprintf("%d", id)
}

// stringToID converts a numeric string to an uint, panicking if the conversion fails.
func stringToID(id string) uint {
	val, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		panic(err)
	}
	return uint(val)
}
