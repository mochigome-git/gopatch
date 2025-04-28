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

func TestCompareAndUpdateNestedMap(t *testing.T) {
	tests := []struct {
		name         string
		parentMap    map[string]map[string]interface{}
		parentKey    string
		updateData   map[string]interface{}
		keysToCheck  []string
		initialPrev  float64
		expectedMap  map[string]map[string]interface{}
		expectedPrev float64
	}{
		{
			name: "Updates larger values only",
			parentMap: map[string]map[string]interface{}{
				"machine1": {
					"ch1_weighing": 100.0,
					"ch2_weighing": 200.0,
					"ch3_weighing": 150.0,
				},
			},
			parentKey: "machine1",
			updateData: map[string]interface{}{
				"ch1_weighing": 120.0, // Should update
				"ch2_weighing": 180.0, // Should not update
				"ch3_weighing": 0.0,   // Should be ignored
			},
			keysToCheck: []string{"ch1_weighing", "ch2_weighing", "ch3_weighing"},
			initialPrev: 110.0,
			expectedMap: map[string]map[string]interface{}{
				"machine1": {
					"ch1_weighing": 120.0,
					"ch2_weighing": 200.0,
					"ch3_weighing": 150.0,
				},
			},
			expectedPrev: 120.0,
		},
		{
			name: "No updates because all new values are smaller",
			parentMap: map[string]map[string]interface{}{
				"machine2": {
					"ch1_weighing": 300.0,
					"ch2_weighing": 400.0,
				},
			},
			parentKey: "machine2",
			updateData: map[string]interface{}{
				"ch1_weighing": 250.0,
				"ch2_weighing": 350.0,
			},
			keysToCheck: []string{"ch1_weighing", "ch2_weighing"},
			initialPrev: 100.0,
			expectedMap: map[string]map[string]interface{}{
				"machine2": {
					"ch1_weighing": 300.0,
					"ch2_weighing": 400.0,
				},
			},
			expectedPrev: 100.0,
		},
		{
			name: "Skip non-float values and zeros",
			parentMap: map[string]map[string]interface{}{
				"machine3": {
					"ch1_weighing": 50.0,
				},
			},
			parentKey: "machine3",
			updateData: map[string]interface{}{
				"ch1_weighing": 0.0,
				"ch2_weighing": "invalid",
			},
			keysToCheck: []string{"ch1_weighing", "ch2_weighing"},
			initialPrev: 10.0,
			expectedMap: map[string]map[string]interface{}{
				"machine3": {
					"ch1_weighing": 50.0,
				},
			},
			expectedPrev: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prev := tt.initialPrev

			CompareAndUpdateNestedMap(tt.parentMap, tt.parentKey, tt.updateData, tt.keysToCheck, &prev)

			if !reflect.DeepEqual(tt.parentMap, tt.expectedMap) {
				t.Errorf("parentMap mismatch. Got %+v, want %+v", tt.parentMap, tt.expectedMap)
			}

			if prev != tt.expectedPrev {
				t.Errorf("prevWeightValue mismatch. Got %v, want %v", prev, tt.expectedPrev)
			}
		})
	}
}
