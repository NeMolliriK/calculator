package global

type ExpressionDTO struct {
	ID     string
	UserID uint
	Data   string
	Status string
	Result float64
}

type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type Result struct {
	value float64
}

type Future struct {
	result chan Result
}

func NewFuture() *Future {
	return &Future{make(chan Result, 1)}
}

func (f *Future) SetResult(val float64) {
	f.result <- Result{val}
}

func (f *Future) Get() float64 {
	res := <-f.result
	return res.value
}
