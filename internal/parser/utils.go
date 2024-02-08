package parser

import "strconv"

// parseInt transforms a string into an int.
func parseInt(in string) int64 {
	v, _ := strconv.ParseInt(in, 10, 64)
	return v
}
