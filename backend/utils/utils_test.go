package utils

import (
	"reflect"
	"testing"
)

func TestSliceWithoutElement_RemoveMiddle(t *testing.T) {
	original := []string{"a", "b", "c", "d", "e"}
	expected := []string{"a", "b", "d", "e"}
	result := SliceWithoutElement(original, 2)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestSliceWithoutElement_RemoveEnd(t *testing.T) {
	original := []string{"a", "b", "c", "d", "e"}
	expected := []string{"a", "b", "c", "d"}
	result := SliceWithoutElement(original, 4)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestSliceWithoutElement_RemoveBeginning(t *testing.T) {
	original := []string{"a", "b", "c", "d", "e"}
	expected := []string{"b", "c", "d", "e"}
	result := SliceWithoutElement(original, 0)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
