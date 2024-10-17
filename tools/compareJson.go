package tools

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
				if _, ok := v.(string); ok {
					continue
				}
				DeleteFields(v.(map[string]interface{}), fields...)
			}
		}
	}
}
