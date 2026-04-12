package collections

func GetOrDefault[K comparable, V any](data map[K]V, key K, defaultValue V) V {
	if value, found := data[key]; found {
		return value
	}
	return defaultValue
}
