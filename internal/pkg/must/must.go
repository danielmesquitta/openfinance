package must

// Must panics if err is not nil, otherwise returns value.
// Use for functions that return (T, error).
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}

	return value
}

// Must0 panics if err is not nil.
// Use for functions that return only error.
func Must0(err error) {
	if err != nil {
		panic(err)
	}
}

// Must2 panics if err is not nil, otherwise returns both values.
// Use for functions that return (T1, T2, error).
func Must2[T1, T2 any](v1 T1, v2 T2, err error) (T1, T2) {
	if err != nil {
		panic(err)
	}

	return v1, v2
}

// Must3 panics if err is not nil, otherwise returns all three values.
// Use for functions that return (T1, T2, T3, error).
func Must3[T1, T2, T3 any](v1 T1, v2 T2, v3 T3, err error) (T1, T2, T3) {
	if err != nil {
		panic(err)
	}

	return v1, v2, v3
}
