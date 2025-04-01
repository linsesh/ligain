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

func TestMapKeys(t *testing.T) {
	// Test with empty map
	emptyMap := make(map[string]int)
	emptyKeys := MapKeys(emptyMap)
	if len(emptyKeys) != 0 {
		t.Error("MapKeys should return empty slice for empty map")
	}

	// Test with string-int map
	stringIntMap := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	keys := MapKeys(stringIntMap)
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Create a map to check if all keys are present
	expectedKeys := map[string]bool{
		"one":   true,
		"two":   true,
		"three": true,
	}
	for _, key := range keys {
		if !expectedKeys[key] {
			t.Errorf("Unexpected key found: %s", key)
		}
	}

	// Test with int-string map
	intStringMap := map[int]string{
		1: "one",
		2: "two",
		3: "three",
	}
	intKeys := MapKeys(intStringMap)
	if len(intKeys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(intKeys))
	}

	// Create a map to check if all int keys are present
	expectedIntKeys := map[int]bool{
		1: true,
		2: true,
		3: true,
	}
	for _, key := range intKeys {
		if !expectedIntKeys[key] {
			t.Errorf("Unexpected key found: %d", key)
		}
	}
}

func TestMapValues(t *testing.T) {
	// Test with empty map
	emptyMap := make(map[string]int)
	emptyValues := MapValues(emptyMap)
	if len(emptyValues) != 0 {
		t.Error("MapValues should return empty slice for empty map")
	}

	// Test with string-int map
	stringIntMap := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	values := MapValues(stringIntMap)
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}

	// Create a map to check if all values are present
	expectedValues := map[int]bool{
		1: true,
		2: true,
		3: true,
	}
	for _, value := range values {
		if !expectedValues[value] {
			t.Errorf("Unexpected value found: %d", value)
		}
	}

	// Test with int-string map
	intStringMap := map[int]string{
		1: "one",
		2: "two",
		3: "three",
	}
	stringValues := MapValues(intStringMap)
	if len(stringValues) != 3 {
		t.Errorf("Expected 3 values, got %d", len(stringValues))
	}

	// Create a map to check if all string values are present
	expectedStringValues := map[string]bool{
		"one":   true,
		"two":   true,
		"three": true,
	}
	for _, value := range stringValues {
		if !expectedStringValues[value] {
			t.Errorf("Unexpected value found: %s", value)
		}
	}

	// Test with duplicate values
	duplicateMap := map[string]int{
		"a": 1,
		"b": 1,
		"c": 2,
	}
	duplicateValues := MapValues(duplicateMap)
	if len(duplicateValues) != 3 {
		t.Errorf("Expected 3 values (including duplicates), got %d", len(duplicateValues))
	}

	// Count occurrences of each value
	valueCounts := make(map[int]int)
	for _, value := range duplicateValues {
		valueCounts[value]++
	}
	if valueCounts[1] != 2 {
		t.Errorf("Expected value 1 to appear twice, got %d times", valueCounts[1])
	}
	if valueCounts[2] != 1 {
		t.Errorf("Expected value 2 to appear once, got %d times", valueCounts[2])
	}
}

func TestMapKeysValues(t *testing.T) {
	// Test with empty map
	emptyMap := make(map[string]int)
	emptyKeys, emptyValues := MapKeysValues(emptyMap)
	if len(emptyKeys) != 0 || len(emptyValues) != 0 {
		t.Error("MapKeysValues should return empty slices for empty map")
	}

	// Test with string-int map
	stringIntMap := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	keys, values := MapKeysValues(stringIntMap)
	if len(keys) != 3 || len(values) != 3 {
		t.Errorf("Expected slices of length 3, got keys: %d, values: %d", len(keys), len(values))
	}

	// Create maps to check if all keys and values are present
	expectedKeys := map[string]bool{
		"one":   true,
		"two":   true,
		"three": true,
	}
	expectedValues := map[int]bool{
		1: true,
		2: true,
		3: true,
	}

	for _, key := range keys {
		if !expectedKeys[key] {
			t.Errorf("Unexpected key found: %s", key)
		}
	}
	for _, value := range values {
		if !expectedValues[value] {
			t.Errorf("Unexpected value found: %d", value)
		}
	}

	// Test with int-string map
	intStringMap := map[int]string{
		1: "one",
		2: "two",
		3: "three",
	}
	intKeys, stringValues := MapKeysValues(intStringMap)
	if len(intKeys) != 3 || len(stringValues) != 3 {
		t.Errorf("Expected slices of length 3, got keys: %d, values: %d", len(intKeys), len(stringValues))
	}

	// Create maps to check if all int keys and string values are present
	expectedIntKeys := map[int]bool{
		1: true,
		2: true,
		3: true,
	}
	expectedStringValues := map[string]bool{
		"one":   true,
		"two":   true,
		"three": true,
	}

	for _, key := range intKeys {
		if !expectedIntKeys[key] {
			t.Errorf("Unexpected key found: %d", key)
		}
	}
	for _, value := range stringValues {
		if !expectedStringValues[value] {
			t.Errorf("Unexpected value found: %s", value)
		}
	}

	// Test with duplicate values
	duplicateMap := map[string]int{
		"a": 1,
		"b": 1,
		"c": 2,
	}
	dupKeys, dupValues := MapKeysValues(duplicateMap)
	if len(dupKeys) != 3 || len(dupValues) != 3 {
		t.Errorf("Expected slices of length 3, got keys: %d, values: %d", len(dupKeys), len(dupValues))
	}

	// Count occurrences of each value
	valueCounts := make(map[int]int)
	for _, value := range dupValues {
		valueCounts[value]++
	}
	if valueCounts[1] != 2 {
		t.Errorf("Expected value 1 to appear twice, got %d times", valueCounts[1])
	}
	if valueCounts[2] != 1 {
		t.Errorf("Expected value 2 to appear once, got %d times", valueCounts[2])
	}
}
