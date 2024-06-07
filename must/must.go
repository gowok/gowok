package must

func Must[T any](res T, err error) T {
	if err != nil {
		panic(err)
	}

	return res
}
