package structures

type Expression struct {
	ID     string
	Data   string
	Status string
	Result float64
}

type ResultWithError struct {
	Value float64
	Err   error
}
