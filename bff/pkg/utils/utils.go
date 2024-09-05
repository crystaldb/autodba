package utils

func MergeMaps[K comparable, V any](map1, map2 map[K]V) map[K]V {
	merged := make(map[K]V, len(map1)+len(map2))
	for key, value := range map1 {
		merged[key] = value
	}
	for key, value := range map2 {
		merged[key] = value
	}
	return merged
}
