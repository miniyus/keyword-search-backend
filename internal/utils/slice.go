package utils

func Map[T interface{}](s []T, fn func(v T, i int) T) []T {
	var mapped []T

	for i, v := range s {
		mapped = append(mapped, fn(v, i))
	}

	return mapped
}

func Filter[T interface{}](s []T, fn func(v T, i int) bool) []T {
	var filtered []T

	for i, v := range s {
		f := fn(v, i)
		if f {
			filtered = append(filtered, v)
		}
	}

	return filtered
}

func Except[T interface{}](s []T, fn func(v T, i int) bool) []T {
	var excepted []T

	for i, v := range s {
		f := fn(v, i)
		if !f {
			excepted = append(excepted, v)
		}
	}

	return excepted
}

func Chunk[T interface{}](s []T, chunkSize int, fn func(v []T, i int)) []T {
	chunkSlice := make([]T, 0)

	for i, v := range s {
		chunkSlice = append(chunkSlice, v)

		if chunkSize == len(chunkSlice) {
			fn(chunkSlice, i)
			chunkSlice = make([]T, 0)
		}
	}

	return s
}

func For[T interface{}](s []T, fn func(v T, i int)) []T {
	for i, v := range s {
		fn(v, i)
	}

	return s
}

func Remove[T interface{}](s []T, index int) []T {
	return append(s[:index], s[index+1:]...)
}
