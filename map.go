package tiga

func MergeMapInterface(src map[string]interface{}, target map[string]interface{}) map[string]interface{} {
	for key, value := range src {
		target[key] = value
	}
	return target
}
