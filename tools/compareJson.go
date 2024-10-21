package tools

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/wI2L/jsondiff"
)

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func DeleteFields(data map[string]interface{}, fields ...string) {
	for key, value := range data {
		if Contains(fields, key) {
			delete(data, key)
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			DeleteFields(nestedMap, fields...)
		} else if nestedArray, ok := value.([]interface{}); ok {
			for _, v := range nestedArray {
				if _, ok := v.(map[string]interface{}); ok {
					DeleteFields(v.(map[string]interface{}), fields...)
				}
			}
		}
	}
}

func IsEqualJson(expected any, res any, logger *Logger, ignoreFields ...string) error {

	if reflect.TypeOf(expected).Kind() != reflect.TypeOf(res).Kind() {
		return fmt.Errorf("differet type of expected result, want: %s, got: %s", reflect.TypeOf(expected).Kind(), reflect.TypeOf(res).Kind())
	}

	if want, ok := expected.(map[string]interface{}); ok {
		err := VerifyKeys(want, res.(map[string]interface{}), "")
		if err != nil {
			return fmt.Errorf("failed to verify keys: %s", err)
		}
	}

	err := removeIgnoredFields(expected, res, ignoreFields...)
	if err != nil {
		return fmt.Errorf("failed to remove ignored fields: %s", err)
	}
	expectedJson, err := json.Marshal(expected)
	if err != nil {
		return fmt.Errorf("failed to marshal expected: %s", err)
	}

	got, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %s", err)
	}

	patch, err := jsondiff.CompareJSON(expectedJson, got)
	if err != nil {
		return fmt.Errorf("failed to compare JSON: %s", err)
	}
	if len(patch.String()) == 0 {
		return nil
	}
	for _, op := range patch {
		logger.Printf("Difference in result: %s\n", op)
	}
	return fmt.Errorf("result is not as expected")
}

func removeIgnoredFields(expected any, res any, ignoreFields ...string) error {

	// expected is string
	if reflect.TypeOf(expected).Kind() == reflect.String {
		return nil
	}

	if reflect.TypeOf(expected).Kind() == reflect.Slice {

		var m map[string]interface{}

		for i := range res.([]any) {

			if reflect.TypeOf(expected.([]any)[i]).Kind() != reflect.TypeOf(m).Kind() ||
				reflect.TypeOf(res.([]any)[i]).Kind() != reflect.TypeOf(m).Kind() {

				return fmt.Errorf("not comparable types")
			}

			// check all keys are present in both objects

			// remove ignored fields
			DeleteFields(res.([]any)[i].(map[string]interface{}), ignoreFields...)
			DeleteFields(expected.([]any)[i].(map[string]interface{}), ignoreFields...)
		}
	} else if reflect.TypeOf(expected).Kind() == reflect.Map {

		DeleteFields(res.(map[string]interface{}), ignoreFields...)
		DeleteFields(expected.(map[string]interface{}), ignoreFields...)
	}

	return nil
}

func VerifyKeys(obj1, obj2 map[string]interface{}, prefix string) error {
	for key, val1 := range obj1 {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		val2, ok := obj2[key]
		if !ok {
			return fmt.Errorf("key '%s' is missing in the result object", fullKey)
		}

		switch v1 := val1.(type) {
		case map[string]interface{}:
			if v2, ok := val2.(map[string]interface{}); ok {
				if err := VerifyKeys(v1, v2, fullKey); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("key '%s' is not an object in the result object", fullKey)
			}
		}
	}
	return nil
}
