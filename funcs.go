// Package demo provides ...
package demo

import (
	"fmt"
	"strconv"
)

func intToString(i int) string {
	var dst []byte
	strconv.AppendInt(dst, int64(i), 10)
	return string(dst)
}
