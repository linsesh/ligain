package utils

type TaskResult[T any] struct {
	Value T
	Err   error
}

func NewTaskResult[T any](value T, err error) *TaskResult[T] {
	return &TaskResult[T]{
		Value: value,
		Err:   err,
	}
}
