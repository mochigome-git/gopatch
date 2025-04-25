package utils

import (
	"os"
	"reflect"
	"testing"
)

// ---- Test for reverseString ----
func TestReverseString(t *testing.T) {
	input := "abc123"
	expected := "321cba"

	result := reverseString(input)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// ---- Test for GetKeyTransformationsFromEnv ----
func TestGetKeyTransformationsFromEnv(t *testing.T) {
	os.Setenv("KEY_TRANSFORMATION_TEST", "d100")
	os.Setenv("KEY_TRANSFORMATION_EXAMPLE", "d101")

	result := GetKeyTransformationsFromEnv("KEY_TRANSFORMATION_")

	expected := map[string]string{
		"TEST":    "d100",
		"EXAMPLE": "d101",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// ---- Test for MergeNonEmptyMaps ----
func TestMergeNonEmptyMaps(t *testing.T) {
	map1 := map[string]interface{}{"a": 1}
	map2 := map[string]interface{}{"b": 2}
	map3 := map[string]interface{}{}

	result := MergeNonEmptyMaps(map1, map2, map3)
	expected := map[string]interface{}{"a": 1, "b": 2}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// ---- Test for parseTriggerKey ----
func TestParseTriggerKey(t *testing.T) {
	input := "trigger1,4,trigger2,7"
	result := parseTriggerKey(input)
	expected := []TriggerKey{
		{"trigger1", "4"},
		{"trigger2", "7"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
