package tools

import (
	"reflect"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{"empty slice", []string{}, "test", false},
		{"single element match", []string{"test"}, "test", true},
		{"single element no match", []string{"test"}, "other", false},
		{"multiple elements match", []string{"test", "other"}, "test", true},
		{"multiple elements no match", []string{"test", "other"}, "another", false},
		{"duplicate elements", []string{"test", "test", "other"}, "test", true},
		{"empty string element", []string{"", "test"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Contains(tt.slice, tt.item)
			if actual != tt.expected {
				t.Errorf("Contains(%v, %q) = %v, want %v", tt.slice, tt.item, actual, tt.expected)
			}
		})
	}
}

func TestDeleteFields(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		fields   []string
		expected map[string]interface{}
	}{
		{
			name: "delete single field",
			data: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			fields: []string{"a"},
			expected: map[string]interface{}{
				"b": 2,
			},
		},
		{
			name: "delete multiple fields",
			data: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			fields: []string{"a", "c"},
			expected: map[string]interface{}{
				"b": 2,
			},
		},
		{
			name: "delete field from nested map",
			data: map[string]interface{}{
				"a": map[string]interface{}{
					"x": 1,
					"y": 2,
				},
			},
			fields: []string{"x"},
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"y": 2,
				},
			},
		},
		{
			name: "delete field from nested array of maps",
			data: map[string]interface{}{
				"a": []interface{}{
					map[string]interface{}{
						"x": 1,
						"y": 2,
					},
					map[string]interface{}{
						"z": 3,
					},
				},
			},
			fields: []string{"x"},
			expected: map[string]interface{}{
				"a": []interface{}{
					map[string]interface{}{
						"y": 2,
					},
					map[string]interface{}{
						"z": 3,
					},
				},
			},
		},
		{
			name: "ignore strings in nested arrays",
			data: map[string]interface{}{
				"a": []interface{}{
					"hello",
					map[string]interface{}{
						"x": 1,
					},
				},
			},
			fields: []string{"x"},
			expected: map[string]interface{}{
				"a": []interface{}{
					"hello",
					map[string]interface{}{},
				},
			},
		},
		{
			name: "handle empty fields slice",
			data: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			fields: []string{},
			expected: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteFields(tt.data, tt.fields...)
			if !reflect.DeepEqual(tt.data, tt.expected) {
				t.Errorf("DeleteFields() = %v, want %v", tt.data, tt.expected)
			}
		})
	}
}
