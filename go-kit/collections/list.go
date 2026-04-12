package collections

func PluckUnique[T any, R comparable](collection []T, iteratee func(item T) R) []R {
	set := NewHashSet[R]()
	for _, item := range collection {
		set.Add(iteratee(item))
	}
	return set.ToSlice()
}

func IsEmpty[T any](collection []T) bool {
	return len(collection) == 0
}

func IsNotEmpty[T any](collection []T) bool {
	return len(collection) > 0
}

func IsEmptyP[T any](items *[]T) bool {
	return items == nil || len(*items) == 0
}

func IsNotEmptyP[M any](items *[]M) bool {
	return items != nil && len(*items) > 0
}

// SliceToGroups [T, K, V] Grouped by K with values transformed to V
func SliceToGroups[T any, K comparable, V any](collection []T, iteratee func(item T) (K, V)) map[K][]V {
	result := map[K][]V{}

	for _, item := range collection {
		key, value := iteratee(item)
		result[key] = append(result[key], value)
	}

	return result
}

// MapUniq [T, K, V] Unique by K with values transformed to V
func MapUniq[T any, K comparable, V any](collection []T, iteratee func(item T, index int) (K, V)) []V {
	result := make([]V, 0)
	seen := map[K]bool{}
	for index, item := range collection {
		k, v := iteratee(item, index)
		if _, ok := seen[k]; !ok {
			seen[k] = true
			result = append(result, v)
		}
	}
	return result
}

// Append appends item to the end of the collection without modifying the original
func Append[T any](collection []T, item ...T) []T {
	result := make([]T, 0, len(collection)+len(item))
	result = append(result, collection...)
	result = append(result, item...)
	return result
}

func Merge[T any](collection []T, other ...[]T) []T {
	finalSize := len(collection)
	for _, otherCollection := range other {
		finalSize += len(otherCollection)
	}
	result := make([]T, 0, finalSize)

	result = append(result, collection...)
	for _, item := range other {
		result = append(result, item...)
	}

	return result
}

func First[T any](collection []T, defaultValue T) T {
	if len(collection) > 0 {
		return collection[0]
	}
	return defaultValue
}
