package utils

// SliceWithoutElement returns a new slice with the element at the given index removed, by copying to a new slice
func SliceWithoutElement[T any](slice []T, index int) []T {
	out := make([]T, 0, len(slice)-1)
	out = append(out, slice[:index]...)
	out = append(out, slice[index+1:]...)
	return out
}
