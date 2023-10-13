package drng

func Must[Value any](value Value, err error) Value {
	if err != nil {
		panic(err)
	}
	return value
}
