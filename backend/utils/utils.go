package utils

// SliceWithoutElement returns a new slice with the element at the given index removed, by copying to a new slice
func SliceWithoutElement[T any](slice []T, index int) []T {
	out := make([]T, 0, len(slice)-1)
	out = append(out, slice[:index]...)
	out = append(out, slice[index+1:]...)
	return out
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
