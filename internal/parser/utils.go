package parser

import "strconv"

func parseInt(in string) int64 {
	v, _ := strconv.ParseInt(in, 10, 64)
	return v
}
