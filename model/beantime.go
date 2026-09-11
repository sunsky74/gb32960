package model

import "fmt"

// BeanTime represents a data collection timestamp in wire format.
// Year stores the wire offset value: Year = actualYear - 2000.
// Example: Year=0x01 means year 2001, Year=0x25 means year 2037.
type BeanTime struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second int
}

// String returns a human-readable representation.
func (bt BeanTime) String() string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		bt.Year+2000, bt.Month, bt.Day, bt.Hour, bt.Minute, bt.Second)
}
