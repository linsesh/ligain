package utils

import (
	"fmt"
	"slices"
	"strings"
)

// SliceWithoutElement returns a new slice with the element at the given index removed, by copying to a new slice
func SliceWithoutElementAtIndex[T any](slice []T, index int) []T {
	out := make([]T, 0, len(slice)-1)
	out = append(out, slice[:index]...)
	out = append(out, slice[index+1:]...)
	return out
}

// SliceWithoutEllement finds the first elem in T and remove it from the slice
func SliceWithoutElement[T comparable](slice []T, elem T) []T {
	index := slices.Index(slice, elem)
	if index == -1 {
		return slice
	}
	return SliceWithoutElementAtIndex(slice, index)
}

// MapKeys returns a slice of the keys of the given map
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// MapValues returns a slice of the values of the given map
func MapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func MapKeysValues[K comparable, V any](m map[K]V) ([]K, []V) {
	keys := make([]K, 0, len(m))
	values := make([]V, 0, len(m))
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}

// ConvertIntSliceToStringWithCommas converts a slice of integers to a comma-separated string
func ConvertIntSliceToStringWithCommas(ids []int) string {
	if len(ids) == 0 {
		return ""
	}

	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = fmt.Sprintf("%d", id)
	}
	return strings.Join(result, ",")
}
