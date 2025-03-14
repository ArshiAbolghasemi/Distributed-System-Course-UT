package util

import (
	"bytes"
	"slices"
)

func FindIndex(s [][]byte, target []byte) int {
    for i, val := range s {
        if bytes.Equal(val, target) {
            return i
        }
    }

    return -1
}

func ReplaceElement[T any](s []T, i int, v T) []T {
    if i < 0 || i >= len(s) {
		return s
	}

	s = slices.Delete(s, i, i+1)
	return append(s[:i], append([]T{v}, s[i:]...)...)
}

